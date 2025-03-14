package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/node"
)

func InstallCommand(args []string) error {
	nvmDir := os.Getenv("NVMDIR")
	if nvmDir == "" {
		return errors.New("environment variable NVMDIR not set")
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

	nodeVersionInstallPath := path.Join(nvmDir, "versions", entry.Version)
	if err := os.MkdirAll(nodeVersionInstallPath, 0o755); err != nil {
		return err
	}

	fmt.Printf("Installing Node %s...\n", entry.Version)
	artifact, err := node.DownloadArtifact(entry.Version)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded artifact %s\n", path.Base(artifact.Name))

	defer os.Remove(artifact.Name)

	fmt.Println("Extracting artifact...")

	if err := node.ExtractArtifact(artifact.Name, nodeVersionInstallPath); err != nil {
		return err
	}

	fmt.Println("Artifact extracted to", nodeVersionInstallPath)
	fmt.Println("Linking additional executables...")

	if err := os.Symlink(path.Join(nodeVersionInstallPath, "lib/node_modules/npm/bin/npm"), path.Join(nodeVersionInstallPath, "bin/npm")); err != nil {
		return err
	}

	if err := os.Symlink(path.Join(nodeVersionInstallPath, "lib/node_modules/npm/bin/npx"), path.Join(nodeVersionInstallPath, "bin/npx")); err != nil {
		return err
	}

	if err := os.Symlink(path.Join(nodeVersionInstallPath, "lib/node_modules/corepack/dist/corepack.js"), path.Join(nodeVersionInstallPath, "bin/corepack")); err != nil {
		return err
	}

	// TODO: set node version to downloaded version

	return nil
}

func InstallCommandUsage() string {
	return `Usage: nvm install [version]

Arguments:
  version (optional)  The version of Node you want to install`
}
