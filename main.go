package main

import (
	"cloudip/cmd"
	"cloudip/util"
	"os"
)

func main() {
	util.EnsureAppDir()

	if err := cmd.Execute(); err != nil {
		util.PrintErrorTrace(err)
		os.Exit(1)
	}
}
