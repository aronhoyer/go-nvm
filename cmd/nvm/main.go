package main

import (
	"fmt"
	"os"
	"path"
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
		Usage:       "install [VERSION] [OPTIONS]",
		Flags: []cli.Flag{
			cli.NewBoolFlagP("use", "u", false, "Activate installed version after install"),
		},
		Run: func(args cli.Args, flags map[string]cli.Flag) error {
			idx, err := node.GetRemoteIndex()
			if err != nil {
				return fmt.Errorf("%w: unable to retrieve node distribution index: %s", cli.ExitCodeUnavailable, err)
			}

			var entry *node.IndexEntry

			switch v := args.Get(0); v {
			case "", "latest":
				entry = &(idx[0])
			case "lts":
				for _, e := range idx {
					if e.LTS != "" {
						entry = &e
						break
					}
				}
			default:
				if !strings.HasPrefix(v, "v") {
					v = "v" + v
				}

				for _, e := range idx {
					if strings.HasPrefix(e.Version, v) {
						entry = &e
						break
					}
				}
			}

			if entry == nil {
				return fmt.Errorf("%w: no such version: %s", cli.ExitCodeUsage, args.Get(0))
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

			if len(idx) == 0 || flags["use"].Value().Get().(bool) {
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
		Usage:       "(rm|remove) <VERSION>",
		Run: func(args cli.Args, flags map[string]cli.Flag) error {
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

	c.Exec()
}
