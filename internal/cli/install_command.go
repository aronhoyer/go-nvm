package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/env"
	"github.com/aronhoyer/go-nvm/internal/node"
)

func InstallCommand(args []string) error {
	nvmDir := os.Getenv("NVMDIR")
	if nvmDir == "" {
		return errors.New("environment variable NVMDIR not set")
	}

	var useInstalledVersion bool

	for i, arg := range args {
		if arg == "-u" || arg == "--use" {
			args = append(args[:i], args[max(i+1, len(args)-1):]...)
			useInstalledVersion = true
			break
		}
	}

	idx, err := node.GetRemoteIndex()
	if err != nil {
		return err
	}

	var entry *node.IndexEntry

	if len(args) == 0 || args[0] == "current" {
		entry = &idx[0]
	} else {
		switch args[0] {
		case "lts":
			// install latest lts
			for _, e := range idx {
				if e.LTS != "" {
					entry = &e
					break
				}
			}
		default:
			// linear search because it's (probably) more likely than not that you'd want to install a version
			// closer to head than tail
			for _, e := range idx {
				if strings.HasPrefix(e.Version, "v"+strings.TrimPrefix(args[0], "v")) {
					entry = &e
					break
				}
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

	dirEntries, _ := os.ReadDir(path.Join(nvmDir, "versions"))

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
