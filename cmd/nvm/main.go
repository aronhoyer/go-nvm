package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/aronhoyer/go-nvm/internal/cli"
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
	var (
		helpFlag    bool
		versionFlag bool
	)

	flag.BoolVar(&helpFlag, "help", false, "")
	flag.BoolVar(&helpFlag, "h", false, "")

	flag.BoolVar(&versionFlag, "version", false, "")
	flag.BoolVar(&versionFlag, "V", false, "")

	flag.Parse()

	c := cli.New(nvmDirPath)

	flag.Usage = func() {
		fmt.Println(cli.Usage())
	}

	if helpFlag {
		flag.Usage()
		return
	}

	if versionFlag {
		fmt.Println(VERSION)
		return
	}

	switch flag.Arg(0) {
	case "version":
		fmt.Println(VERSION)
	case "help":
		s := flag.NewFlagSet("help", flag.ExitOnError)
		s.Parse(flag.Args()[1:])

		if u, err := cli.UsageOf(s.Arg(0)); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			flag.Usage()
			os.Exit(1)
		} else {
			fmt.Println(u)
		}
	case "i", "install":
		if err := c.InstallCommand(flag.Args()[1:]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "use":
		if len(flag.Args()) > 1 {
			switch flag.Arg(1) {
			case "help", "-h", "--help":
				fmt.Println(cli.UseCommandUsage())
			default:
				if err := cli.UseCommand(flag.Args()[1:]); err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, "Error: risetnrisetnrisetrs")
			os.Exit(1)
		}
	case "ls":
		if len(flag.Args()) > 1 {
			switch flag.Arg(1) {
			case "help", "-h", "--help":
				fmt.Println(cli.ListCommandUsage())
			default:
				if err := cli.ListCommand(flag.Args()[1:]); err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}
			}
		} else {
			if err := cli.ListCommand(nil); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported command %s\n", flag.Arg(0))
		fmt.Println(cli.Usage())
		os.Exit(1)
	}
}
