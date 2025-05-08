package platform

import (
	"os"
	"os/exec"
)

func HasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func HasXZSupport(hostOS string) bool {
	if hostOS == "windows" {
		return false
	}

	if hostOS == "freebsd" {
		// check if /usr/lib/liblzma.so exists (freebsd without this file doesn't support xz)
		_, err := os.Stat("/usr/lib/liblzma.so")
		if err == nil {
			return true
		}
	}

	return HasCommand("xz")
}
