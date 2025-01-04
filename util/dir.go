package util

import (
	"cloudip/common"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

func GetAppDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "Cannot get home directory"))
		os.Exit(1)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", common.AppName)
	case "linux":
		return filepath.Join(home, "."+common.AppName)
	}
	PrintErrorTrace(errors.New("unsupported OS"))
	os.Exit(1)
	return ""
}

func EnsureAppDir() {
	// Create the application directory if it doesn't exist
	appDir := GetAppDir()
	ensureDir(appDir)
}

func ensureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			PrintErrorTrace(err)
			os.Exit(1)
		}
	}
}

func IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false // Not exist
	}
	return !info.IsDir() // Exist and not a directory
}
