package node

import (
	"errors"
	"io/fs"
	"os"
	"path"
)

func VersionIsInstalled(version string) (bool, error) {
	nvmDir := os.Getenv("NVMDIR")
	if nvmDir == "" {
		return false, errors.New("environment variable NVMDIR not set")
	}

	_, err := os.Stat(path.Join(nvmDir, "versions", version))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return false, nil
}
