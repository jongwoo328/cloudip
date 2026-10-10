package common

import (
	"fmt"
	"os"
)

var verbose bool

func SetVerbose(v bool) {
	verbose = v
}

func VerboseOutput(msg string) {
	if verbose {
		fmt.Fprintln(os.Stderr, msg)
	}
}
