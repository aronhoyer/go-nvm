package platform

import "os"

func SymlinkForce(oldname, newname string) error {
	if _, err := os.Lstat(newname); err == nil {
		if err := os.Remove(newname); err != nil {
			return err
		}
	}

	if err := os.Symlink(oldname, newname); err != nil {
		return err
	}

	return nil
}
