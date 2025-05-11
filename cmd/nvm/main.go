package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/cli"
	"github.com/aronhoyer/go-nvm/internal/node"
)

const VERSION string = "development (unstable)"

var nvmDirPath string

func init() {
	if nvmDirPath = os.Getenv("NVMDIR"); nvmDirPath == "" {
		if home, err := os.UserHomeDir(); err != nil {
			fmt.Fprintln(os.Stderr, "Error: failed to determine home directory")
			fmt.Println("Try setting the NVMDIR environment variable in your shell's profile")
		} else {
			os.Setenv("NVMDIR", path.Join(home, ".nvm"))
		}
	}
}

func main() {
	c := cli.New(nvmDirPath, &cli.Command{
		Name:        "nvm",
		Description: "Manage Node.js versions",
		Commands: []*cli.Command{
			{
				Name:        "install",
				Aliases:     []string{"i"},
				Description: "Install a Node version",
				Usage:       "install [VERSION] [OPTIONS]",
				Flags: []cli.Flag{
					cli.NewBoolFlagP("use", "u", false, "Activate installed version after install"),
				},
				Run: func(args cli.Args, flags map[string]cli.Flag) error {
					idx, err := node.GetRemoteIndex()
					if err != nil {
						return fmt.Errorf("%w: unable to retrieve node distribution index: %s", cli.ExitCodeUnavailable, err)
					}

					var entry *node.IndexEntry

					switch v := args.Get(0); v {
					case "", "latest":
						entry = &(idx[0])
					case "lts":
						for _, e := range idx {
							if e.LTS != "" {
								entry = &e
								break
							}
						}
					default:
						if !strings.HasPrefix(v, "v") {
							v = "v" + v
						}

						for _, e := range idx {
							if strings.HasPrefix(e.Version, v) {
								entry = &e
								break
							}
						}
					}

					if entry == nil {
						return fmt.Errorf("%w: no such version: %s", cli.ExitCodeUsage, args.Get(0))
					}

					fmt.Println("installing node", entry.Version)

					return nil
				},
			},
			{
				Name:        "remove",
				Aliases:     []string{"rm"},
				Description: "Remove a Node version",
				Usage:       "(rm|remove) <VERSION>",
				Run: func(args cli.Args, flags map[string]cli.Flag) error {
					if args.Get(0) == "" {
						return cli.ExitCodeUsage
					}

					fmt.Println("removing node", args.Get(0))

					return nil
				},
			},
		},
	})

	c.Exec()
}
