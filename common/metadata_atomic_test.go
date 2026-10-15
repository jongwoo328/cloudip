package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetadataWriteFailurePreservesMemory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata")
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	m := &MetadataManager{MetadataFilePath: path, Metadata: &CloudMetadata{Type: AWS, Signature: "old", LastChecked: 42}}
	if err := m.Write(&CloudMetadata{Type: AWS, Signature: "new", LastChecked: 99}); err == nil {
		t.Fatal("expected rename failure")
	}
	if m.Metadata.Signature != "old" || m.Metadata.LastChecked != 42 {
		t.Fatalf("metadata changed after failure: %+v", m.Metadata)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary metadata leaked: %v", entries)
	}
}
