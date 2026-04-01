package main

import (
	"cloudip/cmd"
	"cloudip/common"
	"cloudip/util"
	"os"
)

func main() {
	util.EnsureAppDir(common.AppName)

	if err := cmd.Execute(); err != nil {
		util.PrintErrorTrace(err)
		os.Exit(1)
	}
}
