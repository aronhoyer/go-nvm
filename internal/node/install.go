package node

import (
	"errors"
	"fmt"
	"os"
	"path"
)

func Install(version string) error {
	nvmDir := os.Getenv("NVMDIR")
	if nvmDir == "" {
		return errors.New("environment variable NVMDIR not set")
	}

	fmt.Printf("Installing Node %s...\n", version)

	fmt.Printf("Downloading %s artifact...\n", version)
	artifact, err := DownloadArtifact(version)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded artifact %s\n", path.Base(artifact.Name))

	defer os.Remove(artifact.Name)

	fmt.Println("Extracting artifact...")

	nodeVersionInstallPath := path.Join(nvmDir, "versions", version)

	if err := ExtractArtifact(artifact.Name, nodeVersionInstallPath); err != nil {
		return err
	}

	if err := os.Chmod(path.Join(nodeVersionInstallPath, "bin/node"), 0o744); err != nil {
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

	return nil
}
