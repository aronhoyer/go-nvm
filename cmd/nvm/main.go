package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const VERSION string = "v1.0.0-alpha.0"

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
				if strings.HasPrefix(e.version, "v"+args[1]) {
					entry = &e
					break
				}
			}
		}
	}

	if entry == nil {
		return fmt.Errorf("version %s not found", args[1])
	}

	if err := downloadArtifact(entry.version); err != nil {
		return err
	}

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

func downloadArtifact(v string) error {
	slug, err := getSlug(v)
	if err != nil {
		return err
	}

	u, err := url.JoinPath("https://nodejs.org/dist", v, slug)
	if err != nil {
		return err
	}

	r, err := http.Get(u)
	if err != nil {
		return err
	}

	if r.StatusCode >= 400 {
		return fmt.Errorf("failed to download artifact %s: request failed with status %s", slug, r.Status)
	}

	defer r.Body.Close()

	f, err := os.Create(slug)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return err
	}

	return nil
}
