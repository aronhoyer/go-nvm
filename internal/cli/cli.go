package cli

import "path"

type Cli struct {
	nvmDir, nvmBin, versionsDir string
}

func New(nvmDir string) *Cli {
	return &Cli{
		nvmDir,
		path.Join(nvmDir, "bin"),
		path.Join(nvmDir, "versions"),
	}
}
