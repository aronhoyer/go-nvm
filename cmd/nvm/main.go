package main

import (
	"fmt"
	"os"
	"path"

	"github.com/aronhoyer/go-nvm/internal/cli"
)

const VERSION string = "development (unstable)"

func init() {
	if nvmDir := os.Getenv("NVMDIR"); nvmDir == "" {
		if home, err := os.UserHomeDir(); err != nil {
			fmt.Fprintln(os.Stderr, "Error: failed to determine home directory")
			fmt.Println("Try setting the NVMDIR environment variable in your shell profile")
		} else {
			os.Setenv("NVMDIR", path.Join(home, ".go-nvm"))
		}
	}

	if nvmBin := os.Getenv("NVMBIN"); nvmBin == "" {
		os.Setenv("NVMBIN", path.Join(os.Getenv("NVMDIR"), "bin"))
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		os.Exit(1)
	}

	switch args[0] {
	case "version", "-v", "--version":
		fmt.Println(VERSION)
	case "-h", "--help":
		fmt.Println(cli.Usage())
	case "help":
		if len(args) > 1 {
			if u, err := cli.UsageOf(args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m %v\n\n", err)
				fmt.Println(cli.Usage())
				os.Exit(1)
			} else {
				fmt.Println(u)
			}
		} else {
			fmt.Println(cli.Usage())
		}
	case "i", "install":
		if len(args) > 1 {
			switch args[1] {
			case "help", "-h", "--help":
				fmt.Println(cli.InstallCommandUsage())
			default:
				if err := cli.InstallCommand(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}
			}
		} else {
			if err := cli.InstallCommand(args[1:]); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
		}
	case "use":
		if len(args) > 1 {
			switch args[1] {
			case "help", "-h", "--help":
				fmt.Println(cli.UseCommandUsage())
			default:
				if err := cli.UseCommand(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, "Error: risetnrisetnrisetrs")
			os.Exit(1)
		}
	case "ls":
		if len(args) > 1 {
			switch args[1] {
			case "help", "-h", "--help":
				fmt.Println(cli.ListCommandUsage())
			default:
				if err := cli.ListCommand(args[1:]); err != nil {
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
		fmt.Fprintf(os.Stderr, "Unsupported command %s\n", args[0])
		fmt.Println(cli.Usage())
		os.Exit(1)
	}
}
