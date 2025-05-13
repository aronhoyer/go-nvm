package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

type Value interface {
	String() string
	Set(string) error
	Get() any
}

type boolValue bool

func (v *boolValue) String() string {
	return strconv.FormatBool(bool(*v))
}

func (v *boolValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	*v = boolValue(b)
	return nil
}

func (v *boolValue) Get() any {
	return bool(*v)
}

func newBoolValue(b bool) *boolValue {
	return (*boolValue)(&b)
}

type Flag interface {
	Name() (string, string)
	Description() string
	Value() Value
}

type BoolFlag struct {
	long, short, description string
	value                    Value
}

func (f *BoolFlag) Name() (string, string) {
	return f.long, f.short
}

func (f *BoolFlag) Description() string {
	return f.description
}

func (f *BoolFlag) Value() Value {
	return f.value
}

func NewBoolFlagP(long, short string, defVal bool, description string) Flag {
	return &BoolFlag{long, short, description, newBoolValue(defVal)}
}

type Args []string

func (s *Args) Get(n int) string {
	if n > len(*s)-1 {
		return ""
	}

	return (*s)[n]
}

type Command struct {
	Name        string
	Version     string
	Aliases     []string
	Description string
	Usage       string
	Flags       []Flag
	Commands    []*Command
	Run         func(args Args, flags map[string]Flag) error
}

func (cmd *Command) exec(args []string) {
	var hasHelpFlag bool

	for _, flag := range cmd.Flags {
		long, _ := flag.Name()
		if long == "help" {
			hasHelpFlag = true
			break
		}
	}

	if !hasHelpFlag {
		cmd.Flags = append(cmd.Flags, NewBoolFlagP("help", "h", false, "Print help"))
	}

	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "-h", "--help":
		if cmd.Description != "" {
			fmt.Println(cmd.Description + "\n")
		}
		cmd.printUsage()
		return
	default:
		for _, sub := range cmd.Commands {
			if sub.Name == arg || slices.Contains(sub.Aliases, arg) {
				sub.exec(args[1:])
				return
			}
		}
	}

	flags := make(map[string]Flag)
	var remaining Args

	for _, arg := range args {
		if len(arg) > 1 && arg[0] == '-' {
			var found bool

			for i := range len(cmd.Flags) - 1 {
				flag := cmd.Flags[i]
				long, short := flag.Name()
				if arg == "--"+long || arg == "-"+short {
					switch flag.Value().Get().(type) {
					case bool:
						if err := flag.Value().Set("true"); err != nil {
							fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m invalid boolean argument for -%s, --%s\n", short, long)
							os.Exit(64)
						}

						flags[long] = flag
					}

					found = true
				}
			}

			if !found {
				fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m invalid option: %s\n\n", arg)
				cmd.printUsage()
				os.Exit(ExitCodeUsage.Code())
			}
		}
	}

	for _, flag := range cmd.Flags {
		long, _ := flag.Name()
		if _, ok := flags[long]; !ok {
			flags[long] = flag
		}
	}

	remaining = append(remaining, arg)

	if cmd.Run != nil {
		if err := cmd.Run(args, flags); err != nil {
			fmt.Fprintln(os.Stderr, "\x1b[1;31mError:\x1b[0m", err)
			var exitErr ExitCode
			if errors.As(err, &exitErr) {
				if exitErr == ExitCodeUsage {
					fmt.Print("\n")
					cmd.printUsage()
				}
				os.Exit(exitErr.Code())
			} else {
				os.Exit(ExitCodeSoftware.Code())
			}
		}
	} else if arg != "" {
		fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m invalid command: %s\n\n", arg)
		cmd.printUsage()
		os.Exit(ExitCodeUsage.Code())
	}
}

func (cmd *Command) printUsage() {
	fmt.Print("\x1b[1mUsage:\x1b[0m ")

	if cmd.Usage != "" {
		fmt.Print(cmd.Usage)
	} else {
		fmt.Printf("%s", cmd.Name)

		if len(cmd.Commands) > 0 {
			fmt.Print(" <COMMAND>")
		}
		if len(cmd.Flags) > 0 {
			fmt.Print(" [OPTIONS]")
		}
	}

	fmt.Print("\n")

	if len(cmd.Commands) > 0 {
		fmt.Println("\n\x1b[1mCommands:\x1b[0m")

		maxLen := 0
		names := make([]string, len(cmd.Commands))
		for i, sub := range cmd.Commands {
			fullName := ""

			if len(sub.Aliases) > 0 {
				fullName += strings.Join(sub.Aliases, ", ") + ", "
			}

			fullName += sub.Name

			names[i] = fullName
			if len(fullName) > maxLen {
				maxLen = len(fullName)
			}
		}

		for i, sub := range cmd.Commands {
			fmt.Printf("  %-*s  %s\n", maxLen, names[i], sub.Description)
		}
	}

	if len(cmd.Flags) > 0 {
		fmt.Println("\n\x1b[1mOptions:\x1b[0m")

		maxLen := 0
		names := make([]string, len(cmd.Flags))
		for i, flag := range cmd.Flags {
			long, short := flag.Name()
			nameParts := []string{}
			if short != "" {
				nameParts = append(nameParts, "-"+short)
			}
			nameParts = append(nameParts, "--"+long)
			joined := strings.Join(nameParts, ", ")
			names[i] = joined
			if len(joined) > maxLen {
				maxLen = len(joined)
			}
		}

		for i, flag := range cmd.Flags {
			fmt.Printf("  %-*s  %s\n", maxLen, names[i], flag.Description())
		}
	}
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
	c.RootCmd.Flags = append(c.RootCmd.Flags, NewBoolFlagP("help", "h", false, "Print help"))
	c.RootCmd.Flags = append(c.RootCmd.Flags, NewBoolFlagP("version", "V", false, "Print version"))

	args := os.Args[1:]

	if len(args) == 0 {
		c.RootCmd.printUsage()
		os.Exit(ExitCodeUsage.Code())
		return
	}

	if args[0] == "-V" || args[0] == "--version" {
		fmt.Println(c.RootCmd.Version)
		return
	}

	c.RootCmd.exec(args)
}

func (c *Cli) AddCommand(cmd *Command) {
	c.RootCmd.Commands = append(c.RootCmd.Commands, cmd)
}
