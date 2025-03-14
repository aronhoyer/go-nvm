package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aronhoyer/go-nvm/internal/cli"
	"github.com/aronhoyer/go-nvm/internal/node"
)

const VERSION string = "v1.0.0-alpha.0"

var nvmDir = os.Getenv("NVMDIR")

func init() {
	if nvmDir := os.Getenv("NVMDIR"); nvmDir == "" {
		if home, err := os.UserHomeDir(); err != nil {
			fmt.Fprintln(os.Stderr, "Error: failed to determine home directory")
			fmt.Println("Try setting the NVMDIR environment variable in your shell profile")
		} else {
			os.Setenv("NVMDIR", path.Join(home, ".go-nvm"))
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
		switch args[1] {
		case "help", "-h", "--help":
			fmt.Println(cli.InstallCommandUsage())
		default:
			if err := cli.InstallCommand(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, "Error: ", err)
				os.Exit(1)
			}
		}
	case "use":
		if err := use(args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported command %s\n", args[0])
		fmt.Println(usage())
		os.Exit(1)
	}
}

func use(args []string) error {
	var version string

	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		b, err := os.ReadFile(path.Join(cwd, ".nvmrc"))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("no .nvmrc file found in %s", cwd)
			}

			return err
		}

		version = strings.TrimSpace(string(b))
	}

	idx, err := node.GetRemoteIndex()
	if err != nil {
		return err
	}

	switch v := args[0]; v {
	case "help", "-h", "--help":
		fmt.Println(useUsage())
		return nil
	case "lts":
		for _, e := range idx {
			if e.LTS != "" {
				version = e.Version
				break
			}
		}
	case "current", "latest":
		version = idx[0].Version
	default:
		for _, entry := range idx {
			if strings.HasPrefix(entry.Version, "v"+strings.TrimPrefix(v, "v")) {
				version = entry.Version
				break
			}
		}
	}

	if _, err := os.Stat(path.Join(nvmDir, "versions", version)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("Node %s is not installed. Do you want to install it? [y/N] ", version)

			r := bufio.NewReader(os.Stdin)
			ans, err := r.ReadString('\n')
			if err != nil {
				return err
			}

			if strings.ToLower(strings.TrimSpace(ans)) == "y" {
				return install([]string{version})
			}
		}
	}

	return nil
}

func useUsage() string {
	return `Usage: nvm use [version] [options]

Arguments:
  version (optional)  Use specified Node version or ./.nvmrc if omitted

Options:
  -h, --help  Print help`
}

func usage() string {
	return `Usage: nvm [command] [options]

Commands:
  version  Print nvm version
  install  Install latest or the given version of Node
  use      Specify which version of Node to use
  help     Print this message or the help of the given command

Options:
  -v, --version  Print nvm version
  -h, --help     Print help`
}

func usageOf(s string) (string, error) {
	switch s {
	case "i", "install":
		return installUsage(), nil
	case "use":
		return useUsage(), nil
	default:
		return "", fmt.Errorf("command \"%s\" has no use", s)
	}
}

type indexEntry struct {
	version, npm, lts string
	date              time.Time
}

// TODO: should probably actually sort the []indexEntry
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
		entry, err := parseIndexLine(idxLines[i])
		if err != nil {
			return nil, err
		}

		idx = append(idx, entry)
	}

	return idx, nil
}

func parseIndexLine(line string) (indexEntry, error) {
	// version	date	files	npm	v8	uv	zlib	openssl	modules	lts	security
	parts := strings.Fields(line)

	version, npm, lts := parts[0], parts[3], parts[9]

	date, err := time.Parse("2006-01-02", parts[1])
	if err != nil {
		return indexEntry{}, err
	}

	if npm == "-" {
		npm = ""
	}

	if lts == "-" {
		lts = ""
	}

	return indexEntry{version, npm, lts, date}, nil
}

func install(args []string) error {
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Println(cli.InstallCommandUsage())
		return nil
	}

	return cli.InstallCommand(args[1:])
}

func installUsage() string {
	return `Usage: nvm install [version]

Arguments:
  version (optional)  The version of Node you want to install`
}
