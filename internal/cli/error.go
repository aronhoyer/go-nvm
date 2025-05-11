package cli

// ExitCode is a custom int8 type that represents exit codes.
//
// These exit codes are lifted from the sysexits.h standard and are used to indicate
// specific error conditions in a program's execution. Each exit code corresponds
// to a predefined error message that describes the nature of the failure.
//
// ExitCode implements the [error] interface, providing a descriptive error message
// for each exit code. It also implements the [fmt.Stringer] interface, meaning it
// can be used with formatting functions like fmt.Print or fmt.Sprintf to display
// the associated error message as a string.
type ExitCode int8

const (
	// The command was used incorrectly, e.g., with the wrong number of arguments, a bad flag, bad syntax in a
	// parameter, or whatever.
	ExitCodeUsage ExitCode = iota + 64
	// The input data was incorrect in some way. This should only be used for user's data and not system files.
	ExitCodeDataErr
	// An input file (not a system file) did not exist or was not readable. This could also include errors like "No
	// message" to a mailer (if it cared to catch it).
	ExitCodeNoInput
	// The user specified did not exist. This might be used for mail addresses or remote logins.
	ExitCodeNoUser
	// The host specified did not exist. This is used in mail addresses or network requests.
	ExitCodeNoHost
	// A service is unavailable. This can occur if a support program or file does not exist. This can also be used as a
	// catch-all message when something you wanted to do doesn't work, but you don't know why.
	ExitCodeUnavailable
	// An internal software error has been detected. This should be limited to non-operating system related errors if
	// possible.
	ExitCodeSoftware
	// An operating system error has been detected. This is intended to be used for such things as "cannot fork",
	// "cannot create pipe", or the like. It includes things like getuid returning a user that does not exist in the
	// passwd file.
	ExitCodeOSErr
	// Some system file (e.g., /etc/passwd, /etc/utmp, etc.) does not exist, cannot be opened, or has some sort of
	// error (e.g., syntax error).
	ExitCodeOSFile
	// A (user specified) output file cannot be created.
	ExitCodeCantCreate
	// An error occurred while doing I/O on some file.
	ExitCodeIOErr
	// Temporary failure, indicating something that is not really an error. For example that a mailer could not create
	// a connection, and the request should be reattempted later.
	ExitCodeTempFail
	// The remote system returned something that was "not possible" during a protocol exchange.
	ExitCodeProtocol
	// You did not have sufficient permission to perform the operation. This is not intended for file system problems,
	// which should use ExitCodeNoInput or ExitCodeCantCreate, but rather for higher level permissions.
	ExitCodeNoPerm
	// Something was found in an unconfigured or misconfigured state.
	ExitCodeConfig
)

func (c ExitCode) String() string {
	switch c {
	case ExitCodeUsage:
		return "command line usage error"
	case ExitCodeDataErr:
		return "data format error"
	case ExitCodeNoInput:
		return "cannot open input"
	case ExitCodeNoUser:
		return "addressee unknown"
	case ExitCodeNoHost:
		return "host name unknown"
	case ExitCodeUnavailable:
		return "service unavailable"
	case ExitCodeSoftware:
		return "internal software error"
	case ExitCodeOSErr:
		return "system error"
	case ExitCodeOSFile:
		return "critical OS file missing"
	case ExitCodeCantCreate:
		return "cannot create output file"
	case ExitCodeIOErr:
		return "io failure"
	case ExitCodeTempFail:
		return "temporary failure"
	case ExitCodeProtocol:
		return "remote error in protocol"
	case ExitCodeNoPerm:
		return "permission denied"
	case ExitCodeConfig:
		return "configuration error"
	default:
		return "unknown error"
	}
}

func (c ExitCode) Error() string {
	return c.String()
}

func (c ExitCode) Code() int {
	return int(c)
}
