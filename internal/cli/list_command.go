package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/node"
)

func (cli *Cli) ListCommand(args []string) error {
	var (
		idx []node.IndexEntry
		err error
	)

	if len(args) > 0 {
		switch args[0] {
		case "-r", "--remote":
			idx, err = node.GetRemoteIndex()
		}
	} else {
		idx, err = node.GetLocalIndex(cli.VersionsDirPath())
	}

	if err != nil {
		return err
	}

	printIndex(idx)

	return nil
}

func printIndex(idx []node.IndexEntry) {
	b, _ := exec.Command("node", "--version").Output()
	installedNodeVersion := strings.TrimSpace(string(b))

	for i := len(idx) - 1; i >= 0; i-- {
		e := idx[i]

		if e.Version == installedNodeVersion {
			fmt.Printf("\x1b[32m->%13s", e.Version)
		} else {
			fmt.Printf("%15s", e.Version)
		}

		if e.LTS != "" {
			if e.LTS != idx[max(i-1, 0)].LTS || i == 0 {
				fmt.Printf("\x1b[1;32m   (Latest LTS: %s)\x1b[0m\n", e.LTS)
			} else {
				fmt.Printf("   (LTS: %s)\n", e.LTS)
			}
		} else {
			fmt.Print("\x1b[0m\n")
		}
	}
}

func (cli *Cli) ListCommandUsage() string {
	return `Usage: nvm ls [options]

Options:
  -r, --remote  List remote versions
  -h, --help    Print help`
}
