package cli

import (
	"os"
	"path"
)

type Args []string

func (s *Args) Get(n int) string {
	if n > len(*s)-1 {
		return ""
	}

	return (*s)[n]
}

type Cli struct {
	nvmDir  string
	Version string
	RootCmd *Command
}

func New(nvmDir string, rootCmd *Command) *Cli {
	if rootCmd == nil {
		panic("must provide a root command")
	}

	return &Cli{
		nvmDir:  nvmDir,
		RootCmd: rootCmd,
	}
}

func (c *Cli) RootPath() string {
	return c.nvmDir
}

func (c *Cli) BinPath() string {
	return path.Join(c.nvmDir, "bin")
}

func (c *Cli) VersionsDirPath() string {
	return path.Join(c.nvmDir, "versions")
}

func (c *Cli) Exec() {
	c.RootCmd.exec(os.Args[1:])
}

func (c *Cli) AddCommand(cmd *Command) {
	c.RootCmd.Commands = append(c.RootCmd.Commands, cmd)
}
