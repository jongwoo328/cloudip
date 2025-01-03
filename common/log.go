package common

import "fmt"

func VerboseOutput(msg string) {
	if Flags.Verbose {
		fmt.Println(msg)
	}
}
