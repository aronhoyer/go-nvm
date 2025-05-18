package cli

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var Version func()

type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Flags       []Flag
	Commands    []*Command
	Run         func(args Args, flags FlagSet) error

	parent *Command
}

func (cmd *Command) isRoot() bool {
	return cmd.parent == nil
}

func (cmd *Command) exec(args []string) {
	if cmd.isRoot() {
		if Version != nil {
			cmd.Flags = append(cmd.Flags, NewBoolFlagP("version", "V", false, "Print version"))
		}
		// TODO: could add help command
	}

	cmd.Flags = append(cmd.Flags, NewBoolFlagP("help", "h", false, "Print help"))

	parsedArgs, flags, err := cmd.parseArgs(args)

	if cmd.isRoot() && len(cmd.Commands) > 0 && len(parsedArgs) == 0 {
		fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m a command is required\n\n")
		cmd.printUsage()
		os.Exit(ExitCodeUsage.Code())
	}

	if flags.GetBool("help") {
		if cmd.Description != "" {
			fmt.Println(cmd.Description + "\n")
		}
		cmd.printUsage()
		return
	}

	if flags.GetBool("version") && Version != nil {
		Version()
		return
	}

	for _, sub := range cmd.Commands {
		sub.parent = cmd
		if sub.Name == parsedArgs.Get(0) || slices.Contains(sub.Aliases, parsedArgs.Get(0)) {
			sub.exec(args[1:])
			return
		}
	}

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
		if err := cmd.Run(parsedArgs, flags); err != nil {
			var exErr ExitCode
			if errors.As(err, &exErr) {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(exErr.Code())
			}
		}
	} else if len(parsedArgs) > 0 {
		fmt.Printf("\x1b[1;31mError:\x1b[0m unknown command: %q\n", parsedArgs.Get(0))
		os.Exit(ExitCodeUsage.Code())
	}
}

func (cmd *Command) parseArgs(args []string) (Args, FlagSet, error) {
	var remaining Args
	flags := make(FlagSet)

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			found := false

			for _, f := range cmd.Flags {
				long, short := f.Name()
				if arg == "--"+long || arg == "-"+short {
					switch f.Value().Get().(type) {
					case bool:
						// BoolFlag.Set() calls PaseBool and ParseBool("true") should (tm) never error
						f.Value().Set("true")
					}

					found = true
				}
			}

			if !found {
				return nil, nil, fmt.Errorf("%w: invalid flag: %s", ExitCodeUsage, arg)
			}
		} else {
			remaining = append(remaining, arg)
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
