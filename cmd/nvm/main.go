package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/cli"
	"github.com/aronhoyer/go-nvm/internal/node"
	"github.com/aronhoyer/go-nvm/internal/platform"
)

const VERSION string = "v1.0.0-alpha"

var nvmDirPath string

func init() {
	if nvmDirPath = os.Getenv("NVMDIR"); nvmDirPath == "" {
		if home, err := os.UserHomeDir(); err != nil {
			fmt.Fprintln(os.Stderr, "Error: failed to determine home directory")
			fmt.Println("Try setting the NVMDIR environment variable in your shell's profile")
		} else {
			os.Setenv("NVMDIR", path.Join(home, ".nvm"))
		}
	}

	if err := os.MkdirAll(path.Join(nvmDirPath, "versions"), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "\x1b[1;31mError:\x1b[0m", err)
		os.Exit(cli.ExitCodeIOErr.Code())
	}
}

func main() {
	c := cli.New(nvmDirPath, &cli.Command{
		Name:        "nvm",
		Version:     VERSION,
		Description: "Manage Node.js versions",
	})

	c.AddCommand(&cli.Command{
		Name:        "install",
		Aliases:     []string{"i"},
		Description: "Install a Node version",
		Usage:       "nvm {i,install} [VERSION] [OPTIONS]",
		Flags: []cli.Flag{
			cli.NewBoolFlagP("use", "u", false, "Activate installed version after install"),
		},
		Run: func(args cli.Args, flags cli.FlagSet) error {
			idx, err := node.GetRemoteIndex()
			if err != nil {
				return fmt.Errorf("%w: unable to retrieve node distribution index: %s", cli.ExitCodeUnavailable, err)
			}

			// store index of latest lts versions
			// this gets written on every install, cause i can't be assed to cache bust
			ltsDir := path.Join(c.RootPath(), "lts")
			os.MkdirAll(ltsDir, 0o755)

			latestLTSPath := ""
			writtenLTS := make(map[string]bool)

			for _, e := range idx {
				ltsName := strings.ToLower(e.LTS)
				if _, ok := writtenLTS[ltsName]; e.LTS != "" && !ok {
					p := path.Join(ltsDir, ltsName)
					if err := os.WriteFile(p, []byte(e.Version), 0o644); err != nil {
						return fmt.Errorf("%w: unable to write lts file: %s", cli.ExitCodeIOErr, err)
					}

					if latestLTSPath == "" {
						latestLTSPath = p
					}

					writtenLTS[ltsName] = true
				}
			}

			if err := platform.SymlinkForce(latestLTSPath, path.Join(ltsDir, "latest")); err != nil {
				return fmt.Errorf("%w: unable to symlink latest lts: %s", cli.ExitCodeIOErr, err)
			}

			version := strings.ToLower(args.Get(0))

			switch version {
			case "", "latest":
				version = idx[0].Version
			case "lts":
				b, err := os.ReadFile(latestLTSPath)
				if err != nil {
					return fmt.Errorf("%w: unable to read latest lts: %s", cli.ExitCodeIOErr, err)
				}
				version = string(b)
			default:
				// try reading lts file if we can't parse version
				if _, _, _, err := parseVersion(version); err != nil {
					b, err := os.ReadFile(path.Join(ltsDir, version))
					if err != nil {
						return fmt.Errorf("%w: unable to read lts file: %s", cli.ExitCodeIOErr, err)
					}
					version = string(b)
				} else {
					if !strings.HasPrefix(version, "v") {
						version = "v" + version
					}
				}
			}

			var entry *node.IndexEntry

			// linear search because it's more likely that you'd wanna install a newer version rather than an
			// old version
			for _, e := range idx {
				if versionsMatcheEager(version, e.Version) {
					entry = &e
					break
				}
			}

			if entry == nil {
				return fmt.Errorf("%w: no such version: %s", cli.ExitCodeUsage, version)
			}

			idx, err = node.GetLocalIndex(c.VersionsDirPath())
			if err != nil {
				return fmt.Errorf("%w: unable to read local index", cli.ExitCodeIOErr)
			}

			for _, e := range idx {
				if e.Version == entry.Version {
					return fmt.Errorf("%w: version already installed: %s", cli.ExitCodeUsage, e.Version)
				}
			}

			hostOS, hostArch := platform.SysInfoNorm()
			var artifactExtension string
			switch hostOS {
			case "win":
				artifactExtension = ".zip"
			default:
				if platform.HasXZSupport(hostOS) {
					artifactExtension = ".tar.xz"
				} else {
					artifactExtension = ".tar.gz"
				}
			}

			slug := node.ArtifactSlug(entry.Version, hostOS, hostArch, artifactExtension)

			artifact, err := node.DownloadArtifact(entry.Version, slug)
			if err != nil {
				return fmt.Errorf("%w: failed to download artifact %s", cli.ExitCodeSoftware, slug)
			}

			extractionDst := path.Join(c.VersionsDirPath(), entry.Version)

			if err := os.MkdirAll(extractionDst, 0o755); err != nil {
				return fmt.Errorf("%w: failed to create extraction destination %s", cli.ExitCodeCantCreate, extractionDst)
			}

			if err := node.ExtractArtifact(artifact.Name, extractionDst); err != nil {
				return fmt.Errorf("%w: failed to extract artifact %s", cli.ExitCodeSoftware, artifact.Name)
			}

			if len(idx) == 0 || flags.GetBool("use") {
				if err := os.RemoveAll(c.BinPath()); err != nil {
					return fmt.Errorf("%w: failed to delete %s", cli.ExitCodeIOErr, c.BinPath())
				}

				vbin := path.Join(c.VersionsDirPath(), entry.Version, "bin")
				if err := os.Symlink(vbin, c.BinPath()); err != nil {
					return fmt.Errorf("%w: failed to symlink %s", cli.ExitCodeIOErr, vbin)
				}
			}

			return nil
		},
	})

	c.AddCommand(&cli.Command{
		Name:        "remove",
		Aliases:     []string{"rm"},
		Description: "Remove a Node version",
		Usage:       "nvm {rm,remove} <VERSION>",
		Run: func(args cli.Args, flags cli.FlagSet) error {
			version := args.Get(0)

			if version == "" {
				return cli.ExitCodeUsage
			}

			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}

			idx, err := node.GetLocalIndex(c.VersionsDirPath())
			if err != nil {
				return fmt.Errorf("%w: unable to read local index", cli.ExitCodeIOErr)
			}

			var entry *node.IndexEntry

			for _, e := range idx {
				if strings.HasPrefix(e.Version, version) {
					entry = &e
					break
				}
			}

			if entry == nil {
				return fmt.Errorf("%w: %s: no such version", cli.ExitCodeUsage, version)
			}

			boundVersionPath, err := os.Readlink(c.BinPath())
			if err != nil {
				return fmt.Errorf("%w: unable to read bin link", cli.ExitCodeIOErr)
			}

			versionPath := path.Join(c.VersionsDirPath(), entry.Version)
			if err := os.RemoveAll(versionPath); err != nil {
				return fmt.Errorf("%w: unable to delete %s", cli.ExitCodeIOErr, versionPath)
			}

			if path.Dir(boundVersionPath) == versionPath {
				os.RemoveAll(c.BinPath())
				fmt.Printf("Node %s was active and has been removed. Run `nvm use <VERSION>` to activate another\n", entry.Version)
			}

			return nil
		},
	})

	c.AddCommand(&cli.Command{
		Name:        "use",
		Description: "Activate a version",
		Usage:       "nvm use <VERSION>",
		Run: func(args cli.Args, flags cli.FlagSet) error {
			version := args.Get(0)

			if version == "" {
				wd, err := os.Getwd()
				if err != nil {
					return cli.ExitCodeSoftware
				}

				nvmrc := path.Join(wd, ".nvmrc")
				f, err := os.Open(nvmrc)
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						return fmt.Errorf("%w: no .nvmrc in current directory", cli.ExitCodeUsage)
					}

					return fmt.Errorf("%w: failed to open %s", cli.ExitCodeIOErr, nvmrc)
				}

				b, err := io.ReadAll(f)
				if err != nil {
					return fmt.Errorf("%w: failed to read %s", cli.ExitCodeIOErr, nvmrc)
				}

				version = string(bytes.TrimSpace(b))
			}

			// TODO: check if version already linked?

			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}

			idx, err := node.GetLocalIndex(c.VersionsDirPath())
			if err != nil {
				return fmt.Errorf("%w: failed to read local index", cli.ExitCodeIOErr)
			}

			var entry *node.IndexEntry
			for _, e := range idx {
				if strings.HasPrefix(e.Version, version) {
					entry = &e
					break
				}
			}

			if entry == nil {
				return fmt.Errorf("%w: no such version: %s", cli.ExitCodeUsage, version)
			}

			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}

			if err := os.RemoveAll(c.BinPath()); err != nil {
				return fmt.Errorf("%w: failed to remove existing bin", cli.ExitCodeIOErr)
			}

			if err := os.Symlink(path.Join(c.VersionsDirPath(), entry.Version, "bin"), c.BinPath()); err != nil {
				return fmt.Errorf("%w: failed to symlink version %s", cli.ExitCodeIOErr, version)
			}

			return nil
		},
	})

	c.AddCommand(&cli.Command{
		Name:        "list",
		Aliases:     []string{"ls"},
		Description: "List Node versions",
		Usage:       "nvm {ls,list} [-r,--remote]",
		Flags: []cli.Flag{
			cli.NewBoolFlagP("remote", "r", false, "List versions in remote index"),
		},
		Run: func(args cli.Args, flags cli.FlagSet) error {
			var idx []node.IndexEntry

			if flags.GetBool("remote") {
				ridx, err := node.GetRemoteIndex()
				if err != nil {
					return cli.ExitCodeUnavailable
				}
				idx = ridx
			} else {
				lidx, err := node.GetLocalIndex(c.VersionsDirPath())
				if err != nil {
					return cli.ExitCodeIOErr
				}
				idx = lidx
			}

			activeVersion, err := os.Readlink(c.BinPath())
			if err != nil {
				return cli.ExitCodeIOErr
			}

			activeVersion = path.Base(path.Dir(activeVersion))

			// TODO: some way of mapping local versions to LTS names
			// until then, use lts won't be supported
			for i := len(idx) - 1; i >= 0; i-- {
				entry := idx[i]
				if entry.Version == activeVersion {
					fmt.Printf("\x1b[32m->%13s", entry.Version)
				} else {
					fmt.Printf("%15s", entry.Version)
				}

				isLatestLTS := entry.LTS != "" && idx[max(i-1, 0)].LTS == ""

				if entry.LTS != "" {
					if isLatestLTS {
						fmt.Printf("\x1b[1;32m  (Latest LTS: %s)\x1b[0m", entry.LTS)
					} else {
						fmt.Printf("  (LTS: %s)", entry.LTS)
					}
				}

				fmt.Print("\x1b[0m\n")
			}

			return nil
		},
	})

	c.Exec()
}

var ErrVersionNumber = errors.New("invalid version number")

func parseVersion(v string) (major, minor, patch string, err error) {
	reg := regexp.MustCompile(`\d+(\.\d+)*`)
	if !reg.MatchString(v) {
		err = fmt.Errorf("%w: %s", ErrVersionNumber, v)
		return
	}

	v = strings.TrimPrefix(v, "v")
	p := strings.Split(v, ".")

	if len(p) == 0 || len(p) > 3 {
		err = fmt.Errorf("%w: %s", ErrVersionNumber, v)
		return
	}

	major = p[0]

	if len(p) > 1 {
		minor = p[1]
	}

	if len(p) > 2 {
		patch = p[2]
	}

	return
}

func versionsMatcheEager(a, b string) bool {
	rmaj, rmin, rpatch, _ := parseVersion(a)
	emaj, emin, epatch, _ := parseVersion(b)

	if rmaj != emaj {
		return false
	}

	if rmin != "" && rmin != emin {
		return false
	}

	if rpatch != "" && rpatch != epatch {
		return false
	}

	return true
}
