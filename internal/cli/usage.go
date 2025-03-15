package cli

import "fmt"

func Usage() string {
	return `Usage: nvm [command] [options]

Commands:
  version  Print nvm version
  install  Install latest or the given version of Node
  use      Specify which version of Node to use
  ls       List versions
  help     Print this message or the help of the given command

Options:
  -v, --version  Print nvm version
  -h, --help     Print help`
}

func UsageOf(s string) (string, error) {
	switch s {
	case "i", "install":
		return InstallCommandUsage(), nil
	case "use":
		return UseCommandUsage(), nil
	case "ls":
		return ListCommandUsage(), nil
	default:
		return "", fmt.Errorf("command \"%s\" has no use", s)
	}
}
