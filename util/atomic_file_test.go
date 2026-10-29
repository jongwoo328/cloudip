package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicFilePreservesOriginalOnFailedWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("disk write failed")
	err := WriteFileAtomically(path, func(file *os.File) error {
		if _, err := file.WriteString("partial"); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("error=%v", err)
	}
	assertCacheContent(t, path, "original")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary file leaked: %v", entries)
	}
	if err := WriteFileAtomically(path, func(file *os.File) error { _, err := file.WriteString("new"); return err }); err != nil {
		t.Fatal(err)
	}
	assertCacheContent(t, path, "new")
}

func TestAtomicJSONPreservesOriginalOnEncodingError(t *testing.T) {
	for _, tc := range []struct {
		name string
		data any
	}{
		{"unsupported type", map[string]any{"unsupported": make(chan int)}},
		{"invalid UTF-8", map[string]string{"value": "\xff"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.json")
			if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := WriteJSONFile(path, &tc.data); err == nil {
				t.Fatal("expected encoding error")
			}
			assertCacheContent(t, path, "original")
		})
	}
}

func TestAtomicFileCleansUpOnRenameFailure(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "directory")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	err := WriteFileAtomically(target, func(file *os.File) error { _, err := file.WriteString("new"); return err })
	if err == nil {
		t.Fatal("expected rename failure")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("destination changed or temporary file leaked: %v", entries)
	}
}

func TestInterruptedDownloadPreservesCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.Write([]byte("partial"))
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "cache.json")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := DownloadFromUrlToFileWithHeaders(server.URL, path); err == nil {
		t.Fatal("expected truncated response error")
	}
	assertCacheContent(t, path, "original")
}

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	for _, input := range []string{`{} {}`, `{} garbage`, `{"incomplete":`} {
		var data map[string]any
		if err := DecodeJSON(strings.NewReader(input), &data); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}

func assertCacheContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("cache=%q, want %q", got, want)
	}
}
