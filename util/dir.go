package util

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

func GetAppDir(appName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "Cannot get home directory"))
		os.Exit(1)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	case "linux":
		return filepath.Join(home, "."+appName)
	}
	PrintErrorTrace(errors.New("unsupported OS"))
	os.Exit(1)
	return ""
}

func EnsureAppDir(appName string) {
	appDir := GetAppDir(appName)
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
