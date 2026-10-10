package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func GetAppDir(appName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "cannot get home directory"))
		os.Exit(1)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	case "linux":
		return filepath.Join(home, "."+appName)
	}
	PrintErrorTrace(fmt.Errorf("unsupported OS: %s", runtime.GOOS))
	os.Exit(1)
	return ""
}

func EnsureAppDir(appName string) error {
	appDir := GetAppDir(appName)
	return ensureDir(appDir)
}

func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("cannot create application directory %q: %w", path, err)
	}
	return nil
}

func IsFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cannot stat file %q: %w", path, err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("expected a file at %q, found a directory", path)
	}
	return true, nil
}
