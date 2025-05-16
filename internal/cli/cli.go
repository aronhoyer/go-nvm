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

	parent *Command
}

func (cmd *Command) isRoot() bool {
	return cmd.parent == nil
}

func (cmd *Command) exec(args []string) {
	if cmd.isRoot() {
		cmd.Flags = append(cmd.Flags, NewBoolFlagP("version", "V", false, "Print version"))
		// TODO: could add help command
	}

	cmd.Flags = append(cmd.Flags, NewBoolFlagP("help", "h", false, "Print help"))

	var arg string

	if len(args) > 0 {
		arg = args[0]
	} else {
		if cmd.isRoot() && len(cmd.Commands) > 0 {
			fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m a command is required\n\n")
			cmd.printUsage()
			os.Exit(ExitCodeUsage.Code())
		}
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
			sub.parent = cmd
			if sub.Name == arg || slices.Contains(sub.Aliases, arg) {
				sub.exec(args[1:])
				return
			}
		}
	}

	remainingArgs, flags, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m %s\n\n", err.Error())

		var exErr ExitCode
		if errors.As(err, &exErr) {
			defer os.Exit(exErr.Code())
		} else {
			defer os.Exit(ExitCodeSoftware.Code())
		}

		return
	}

	if cmd.Run != nil {
		if err := cmd.Run(remainingArgs, flags); err != nil {
			var exErr ExitCode
			if errors.As(err, &exErr) {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(exErr.Code())
			}
		}
	} else if remainingArgs.Get(0) == "" {
		fmt.Println("unknown command:", remainingArgs.Get(0))
	}
}

func (cmd *Command) parseArgs(args []string) (Args, map[string]Flag, error) {
	var remaining Args
	flags := make(map[string]Flag)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			remaining = append(remaining, arg)
			args = slices.Concat(args[:i], args[i+1:])
			i++
		}
	}

	for _, arg := range args {
		found := false

		for _, flag := range cmd.Flags {
			long, short := flag.Name()

			if arg == "--"+long || arg == "-"+short {
				found = true

				switch flag.Value().Get().(type) {
				case bool:
					// BoolFlag.Set() calls PaseBool and ParseBool("true") should (tm) never error
					flag.Value().Set("true")
				}
			}
		}

		if !found {
			return nil, nil, fmt.Errorf("%w: invalid flag: %s", ExitCodeUsage, arg)
		}
	}

	for _, flag := range cmd.Flags {
		long, _ := flag.Name()

		if _, ok := flags[long]; !ok {
			flags[long] = flag
		}
	}

	return remaining, flags, nil
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
	c.RootCmd.exec(os.Args[1:])
}

func (c *Cli) AddCommand(cmd *Command) {
	c.RootCmd.Commands = append(c.RootCmd.Commands, cmd)
}
