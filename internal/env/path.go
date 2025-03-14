package env

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
)

func SetNodeVersion(v string) error {
	versionDir := path.Join(os.Getenv("NVMDIR"), "versions", v)

	if _, err := os.Stat(versionDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("node %s not installed", v)
		}

		return err
	}

	if _, err := os.Stat(os.Getenv("NVMBIN")); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	} else {
		if err := os.RemoveAll(os.Getenv("NVMBIN")); err != nil {
			panic(err)
		}
	}

	if err := os.Symlink(path.Join(versionDir, "bin"), os.Getenv("NVMBIN")); err != nil {
		return err
	}

	return nil
}
