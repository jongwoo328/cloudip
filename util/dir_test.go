package util

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestIsFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path      string
		exists, wantErr bool
	}{
		{"file", file, true, false},
		{"missing", filepath.Join(dir, "missing"), false, false},
		{"directory", dir, false, true},
		{"invalid path", "bad\x00path", false, true},
		{"parent is a file", filepath.Join(file, "child"), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := IsFileExists(tc.path)
			if exists != tc.exists || (err != nil) != tc.wantErr {
				t.Fatalf("IsFileExists(%q)=(%v,%v)", tc.path, exists, err)
			}
		})
	}
	_, err := IsFileExists("bad\x00path")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) || !errors.Is(err, syscall.EINVAL) {
		t.Fatalf("stat error not preserved: %v", err)
	}
}

func TestIsFileExistsPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0700) })
	exists, err := IsFileExists(file)
	if exists || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("exists=%v error=%v, want permission error", exists, err)
	}
}

func TestEnsureDirRejectsFileAndCreatesParents(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := ensureDir(file); err == nil {
		t.Fatal("expected error for existing file")
	}
	if err := ensureDir("bad\x00path"); err == nil {
		t.Fatal("expected error for invalid path")
	}
	nested := filepath.Join(dir, "new", "nested")
	if err := ensureDir(nested); err != nil {
		t.Fatal(err)
	}
	if err := ensureDir(nested); err != nil {
		t.Fatal(err)
	}
}
