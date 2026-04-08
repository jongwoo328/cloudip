package common

import "fmt"

var verbose bool

func SetVerbose(v bool) {
	verbose = v
}

func VerboseOutput(msg string) {
	if verbose {
		fmt.Println(msg)
	}
}
