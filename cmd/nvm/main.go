package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	VERSION     string = "v1.0.0-alpha.0"
	NVMDIR_NAME string = ".go-nvm"
)

var nvmDir = os.Getenv("NVMDIR")

func init() {
	if nvmDir == "" {
		if home, err := os.UserHomeDir(); err != nil {
			fmt.Fprintln(os.Stderr, "Error: failed to determine home directory")
			fmt.Println("Try setting the NVMDIR environment variable in your shell profile")
		} else {
			nvmDir = path.Join(home, ".go-nvm")
		}
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		os.Exit(1)
	}

	switch args[0] {
	case "version", "-v", "--version":
		fmt.Println(VERSION)
	case "-h", "--help":
		fmt.Println(usage())
	case "help":
		if len(args) > 1 {
			if u, err := usageOf(args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m %v\n\n", err)
				fmt.Println(usage())
				os.Exit(1)
			} else {
				fmt.Println(u)
			}
		} else {
			fmt.Println(usage())
		}
	case "i", "install":
		if err := install(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported command %s\n", args[0])
		fmt.Println(usage())
		os.Exit(1)
	}
}

func usage() string {
	return `Usage: nvm [command] [options]

Commands:
  version    Print nvm version
  install  Install latest or the given version of Node
  help       Print this message or the help of the given command

Options:
  -v, --version  Print nvm version
  -h, --help     Print help`
}

func usageOf(s string) (string, error) {
	switch s {
	case "i", "install":
		return installUsage(), nil
	default:
		return "", fmt.Errorf("command \"%s\" has no use", s)
	}
}

type indexEntry struct {
	version, npm, lts string
}

func getRemodeIndex() ([]indexEntry, error) {
	res, err := http.Get("https://nodejs.org/dist/index.tab")
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	idxLines := strings.Split(strings.TrimSpace(string(b)), "\n")
	var idx []indexEntry

	for i := 1; i < len(idxLines); i++ {
		idx = append(idx, parseIndexLine(idxLines[i]))
	}

	return idx, nil
}

func parseIndexLine(line string) indexEntry {
	// version	date	files	npm	v8	uv	zlib	openssl	modules	lts	security
	parts := strings.Fields(line)

	version, npm, lts := parts[0], parts[3], parts[9]

	if npm == "-" {
		npm = ""
	}

	if lts == "-" {
		lts = ""
	}

	return indexEntry{version, npm, lts}
}

func install(args []string) error {
	idx, err := getRemodeIndex()
	if err != nil {
		return err
	}

	var entry *indexEntry

	if len(args) == 0 || args[0] == "current" {
		entry = &idx[0]
	} else {
		switch args[0] {
		case "help", "-h", "--help":
			fmt.Println(installUsage())
		case "lts":
			// install latest lts
			for _, e := range idx {
				if e.lts != "" {
					entry = &e
					break
				}
			}
		default:
			// linear search because it's (probably) more likely than not that you'd want to install a version
			// closer to head than tail
			for _, e := range idx {
				if strings.HasPrefix(e.version, "v"+args[0]) {
					entry = &e
					break
				}
			}
		}
	}

	dstDir := path.Join(nvmDir, "versions", entry.version)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		if errors.Is(err, os.ErrExist) {
			// TODO: check if the node version is ACTUALLY installed
			// for now, we just assume so if the directory exists

			return fmt.Errorf("node %s is already installed", entry.version)
		}

		return err
	}

	if entry == nil {
		return fmt.Errorf("version %s not found", args[0])
	}

	fmt.Printf("Installing Node %s...\n", entry.version)
	p, err := downloadArtifact(entry.version)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded artifact %s\n", path.Base(p))

	defer os.Remove(p)

	fmt.Println("Extracting artifact...")

	switch ext := path.Ext(p); ext {
	case ".gz": // just assume .tar.gz
		if err := extractGzipArtifact(p, dstDir); err != nil {
			if err := os.Remove(dstDir); err != nil {
				panic(err)
			}

			return err
		}
	case ".zip":
		if err := extractZipArtifact(p, dstDir); err != nil {
			if err := os.Remove(dstDir); err != nil {
				panic(err)
			}

			return err
		}
	default:
		return fmt.Errorf("compression algorithm %s not supported", ext)
	}

	if err := os.Symlink(path.Join(dstDir, "lib/node_modules/npm/bin/npm"), path.Join(dstDir, "bin/npm")); err != nil {
		return err
	}

	if err := os.Symlink(path.Join(dstDir, "lib/node_modules/npm/bin/npx"), path.Join(dstDir, "bin/npx")); err != nil {
		return err
	}

	if err := os.Symlink(path.Join(dstDir, "lib/node_modules/corepack/dist/corepack.js"), path.Join(dstDir, "bin/corepack")); err != nil {
		return err
	}

	return nil
}

func extractGzipArtifact(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var tld string

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if tld == "" {
			tld = strings.Split(h.Name, "/")[0]
		}

		target := path.Join(dst, strings.TrimPrefix(h.Name, tld+"/"))

		if target == dst {
			continue
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(h.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractZipArtifact(src, dst string) error {
	return nil
}

func installUsage() string {
	return `Usage: nvm install [version]

Arguments:
  version (optional)  The version of Node you want to install`
}

func getSlug(v string) (string, error) {
	var (
		ops  = runtime.GOOS
		arch string
		ext  string
	)

	switch runtime.GOARCH {
	case "386":
		arch = "x86"
	case "amd64":
		arch = "x64"
	case "arm":
		arch = "armv7l"
	}

	switch ops {
	case "aix", "darwin":
		ext = ".tar.gz"
	case "linux":
		switch arch {
		case "arm64", "armv7l", "ppc64le", "s390x", "x64":
			break
		default:
			return "", fmt.Errorf("not supported: %s/%s", ops, arch)
		}

		ext = ".tar.gz"
	case "windows":
		ops = "win"
		ext = ".zip"
	default:
		return "", fmt.Errorf("%s not supported", ops)
	}

	return fmt.Sprintf("node-%s-%s-%s%s", v, ops, arch, ext), nil
}

func downloadArtifact(v string) (string, error) {
	slug, err := getSlug(v)
	if err != nil {
		return "", err
	}

	u, err := url.JoinPath("https://nodejs.org/dist", v, slug)
	if err != nil {
		return "", err
	}

	r, err := http.Get(u)
	if err != nil {
		return "", err
	}

	if r.StatusCode >= 400 {
		return "", fmt.Errorf("failed to download artifact %s: request failed with status %s", slug, r.Status)
	}

	defer r.Body.Close()

	f, err := os.Create(path.Join(os.TempDir(), slug))
	if err != nil {
		return "", err
	}

	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return "", err
	}

	return f.Name(), nil
}
