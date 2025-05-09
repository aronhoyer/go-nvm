package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/env"
	"github.com/aronhoyer/go-nvm/internal/node"
)

func (cli *Cli) InstallCommand(args []string) error {
	s := flag.NewFlagSet("install", flag.ExitOnError)

	var (
		helpFlag            bool
		useInstalledVersion bool
	)

	s.BoolVar(&helpFlag, "help", false, "")
	s.BoolVar(&helpFlag, "h", false, "")

	s.BoolVar(&useInstalledVersion, "use", false, "")
	s.BoolVar(&useInstalledVersion, "u", false, "")

	s.Usage = func() {
		fmt.Println(InstallCommandUsage())
	}

	s.Parse(args)

	if helpFlag {
		s.Usage()
		return nil
	}

	idx, err := node.GetRemoteIndex()
	if err != nil {
		return err
	}

	var entry *node.IndexEntry

	switch v := s.Arg(0); v {
	case "", "current":
		entry = &idx[0]
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
		return fmt.Errorf("version %s not found", args[0])
	}

	isInstalled, err := node.VersionIsInstalled(entry.Version)
	if err != nil {
		return err
	}

	if isInstalled {
		return fmt.Errorf("node %s is already installed", entry.Version)
	}

	if err := node.Install(entry.Version); err != nil {
		return err
	}

	dirEntries, _ := os.ReadDir(cli.nvmVersionsPath)

	if len(dirEntries) == 1 || useInstalledVersion {
		if err := env.SetNodeVersion(entry.Version); err != nil {
			return err
		}
	}

	return nil
}

func InstallCommandUsage() string {
	return `Usage: nvm install [version] [options]

Arguments:
  version (optional)  The version of Node you want to install

Options:
  -u, --use   Use this version after installing. If no other version of Node is installed, this option is implied
  -h, --help  Print help`
}
