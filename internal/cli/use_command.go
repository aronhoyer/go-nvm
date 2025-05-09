package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/env"
	"github.com/aronhoyer/go-nvm/internal/node"
)

func (cli *Cli) UseCommand(args []string) error {
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

	versionInstalled, err := node.VersionIsInstalled(version)
	if err != nil {
		return err
	}

	if !versionInstalled {
		fmt.Printf("Node %s is not installed. Do you want to install it? [y/N] ", version)

		r := bufio.NewReader(os.Stdin)
		ans, err := r.ReadString('\n')
		if err != nil {
			return err
		}

		if strings.ToLower(strings.TrimSpace(ans)) == "y" {
			if err := node.Install(version); err != nil {
				return err
			}
		}
	}

	if err := env.SetNodeVersion(version); err != nil {
		return err
	}

	return nil
}

func (cli *Cli) UseCommandUsage() string {
	return `Usage: nvm use [version] [options]

Arguments:
  version (optional)  Use specified Node version or ./.nvmrc if omitted

Options:
  -h, --help  Print help`
}
