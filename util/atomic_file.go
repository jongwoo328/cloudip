package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomically keeps the previous file intact until the replacement is
// completely written, synced, and closed. The temporary file shares its directory
// so rename is atomic on the supported Unix platforms.
func WriteFileAtomically(path string, write func(*os.File) error) error {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temporary cache: %w", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := file.Chmod(0644); err != nil {
		return fmt.Errorf("set cache permissions: %w", err)
	}
	if err := write(file); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync cache: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close cache: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}
	return nil
}

func WriteJSONFile[T any](path string, data *T) error {
	return WriteFileAtomically(path, func(file *os.File) error { return WriteJSON(file, data) })
}
