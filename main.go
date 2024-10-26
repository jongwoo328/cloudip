package main

import (
	"cloudip/cmd"
	"cloudip/internal"
	"fmt"
	"os"
)

func main() {
	internal.EnsureAppDir()

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
