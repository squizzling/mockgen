package args

import (
	"github.com/jessevdk/go-flags"
)

// IsHelp is a helper to test the error from parseArgs() to
// determine if the help message was written. It is safe to
// call without first checking that err is nil.
func IsHelp(err error) bool {
	// This was copied from https://github.com/jessevdk/go-flags/blame/master/help.go#L499, as there has not been an
	// official release yet with this code. Renamed from WriteHelp to IsHelp, as flags.ErrHelp is still returned when
	// flags.HelpFlag is set, flags.PrintError is clear, and -h/--help is passed on the command line, even though the
	// help is not displayed in such a situation.
	if err == nil { // No error
		return false
	}

	flagError, ok := err.(*flags.Error)
	if !ok { // Not a go-flag error
		return false
	}

	if flagError.Type != flags.ErrHelp { // Did not print the help message
		return false
	}

	return true
}
