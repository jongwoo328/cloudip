package util

import (
	"cloudip/common"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func getAppDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", common.AppName), nil
	case "linux":
		return filepath.Join(home, "."+common.AppName), nil
	default:
		return "", fmt.Errorf("unsupported platform")
	}
}

func GetAppDir() (string, error) {
	return getAppDir()
}

func EnsureAppDir() {
	// Create the application directory if it doesn't exist
	appDir, err := getAppDir()
	if err != nil {
		PrintErrorTrace(err)
		os.Exit(1)
	}
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
