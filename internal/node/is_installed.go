package node

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
)

func VersionIsInstalled(version string) (bool, error) {
	nvmDir := os.Getenv("NVMDIR")
	if nvmDir == "" {
		return false, errors.New("environment variable NVMDIR not set")
	}

	s, err := os.Stat(path.Join(nvmDir, "versions", version))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	if s.IsDir() {
		nodeVersion, _ := exec.Command("node", "--version").Output()
		if strings.TrimSpace(string(nodeVersion)) == version {
			return true, nil
		}
	}

	return false, nil
}
