package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/aronhoyer/go-nvm/internal/node"
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
	Usage       func()
	Flags       []Flag
	Commands    []*Command
	Run         func(args Args, flags map[string]Flag) error
}

func (cmd *Command) Exec(args []string) {
	cmd.Flags = append(cmd.Flags, NewBoolFlagP("help", "h", false, "Print help"))

	if cmd.Usage == nil {
		cmd.Usage = cmd.defaultUsage
	}

	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	switch arg {
	case "-h", "--help":
		fmt.Println(cmd.Description + "\n")
		cmd.Usage()
		return
	default:
		for _, sub := range cmd.Commands {
			if sub.Name == arg || slices.Contains(sub.Aliases, arg) {
				sub.Exec(args[1:])
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
				cmd.Usage()
				os.Exit(64)
			}
		} else {
			for _, flag := range cmd.Flags {
				long, _ := flag.Name()
				if _, ok := flags[long]; !ok {
					flags[long] = flag
				}
			}

			remaining = append(remaining, arg)
		}
	}

	if arg == "" && cmd.Run != nil {
		cmd.Run(args, flags)
	} else if arg != "" {
		fmt.Fprintf(os.Stderr, "\x1b[1;31mError:\x1b[0m invalid command: %s\n\n", arg)
		cmd.Usage()
		os.Exit(64)
	}
}

func (cmd *Command) defaultUsage() {
	fmt.Printf("\x1b[1mUsage:\x1b[0m %s", cmd.Name)

	if len(cmd.Commands) > 0 {
		fmt.Print(" <COMMAND>")
	}
	if len(cmd.Flags) > 0 {
		fmt.Print(" [OPTIONS]")
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

func main() {
	nvm := &Command{
		Name:        "nvm",
		Description: "Manage Node.js versions",
		Commands: []*Command{
			{
				Name:        "install",
				Aliases:     []string{"i"},
				Description: "Install Node version",
				Flags: []Flag{
					NewBoolFlagP("use", "u", false, "Use version after install"),
				},
				Run: func(args Args, flags map[string]Flag) error {
					idx, err := node.GetRemoteIndex()
					if err != nil {
						fmt.Fprintln(os.Stderr, "\x1b[1;31mError:\x1b[0m", err)
						os.Exit(69)
					}

					var version string

					switch v := args.Get(0); v {
					case "", "latest":
						version = idx[0].Version
					case "lts":
						for _, entry := range idx {
							if entry.LTS != "" {
								version = entry.Version
								break
							}
						}
					default:
						if !strings.HasPrefix(v, "v") {
							v = "v" + v
						}

						for _, entry := range idx {
							if strings.HasPrefix(entry.Version, v) {
								version = entry.Version
								break
							}
						}
					}

					fmt.Println("installing Node", version)
					return nil
				},
			},
			{
				Name:        "ls",
				Description: "List version",
				Flags: []Flag{
					NewBoolFlagP("remote", "r", false, "List remote version"),
				},
			},
			{
				Name:        "rm",
				Description: "Uninstall a version",
			},
			{
				Name:        "use",
				Description: "Activate an installed version",
			},
		},
	}

	nvm.Exec(os.Args[1:])
}
