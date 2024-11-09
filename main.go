package main

import (
	"cloudip/cmd"
	"cloudip/internal/util"
	"fmt"
	"os"
)

func main() {
	util.EnsureAppDir()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
