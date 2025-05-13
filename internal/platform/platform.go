package platform

import (
	"os"
	"os/exec"
	"runtime"
)

func SysInfoNorm() (hostOS, hostArch string) {
	hostOS = runtime.GOOS
	hostArch = runtime.GOARCH

	switch hostOS {
	case "solaris":
		hostOS = "sunos"
	case "windows":
		hostOS = "win"
	}

	switch hostArch {
	case "amd64":
		hostArch = "x64"
	case "arm":
		hostArch = "armv7l"
	case "386":
		hostArch = "x86"
	}

	return hostOS, hostArch
}

func HasCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func HasXZSupport(hostOS string) bool {
	if hostOS == "windows" {
		return false
	}

	if hostOS == "freebsd" {
		// check if /usr/lib/liblzma.so exists (freebsd without this file might not support xz)
		_, err := os.Stat("/usr/lib/liblzma.so")
		if err == nil {
			return true
		}
	}

	return HasCommand("xz")
}
