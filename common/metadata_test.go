package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetadataIsSignatureExpired(t *testing.T) {
	manager := &MetadataManager{
		Metadata: &CloudMetadata{
			Type:      AWS,
			Signature: "etag-value",
		},
	}

	if manager.IsSignatureExpired("etag-value") {
		t.Fatal("IsSignatureExpired() = true, want false for matching signature")
	}
	if !manager.IsSignatureExpired("new-etag-value") {
		t.Fatal("IsSignatureExpired() = false, want true for changed signature")
	}
}

func TestMetadataWritePersistsSignatureOnly(t *testing.T) {
	dir := t.TempDir()
	manager := &MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &CloudMetadata{
			Type: AWS,
		},
	}

	if err := manager.Write(&CloudMetadata{Type: AWS, Signature: "new-signature"}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(manager.MetadataFilePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(content), `"signature": "new-signature"`) {
		t.Fatalf("metadata file does not contain new signature: %s", string(content))
	}
	if strings.Contains(string(content), "lastModified") {
		t.Fatalf("metadata file contains legacy lastModified: %s", string(content))
	}
	if manager.Metadata.Signature != "new-signature" {
		t.Fatalf("in-memory signature = %q, want %q", manager.Metadata.Signature, "new-signature")
	}
}

func TestMetadataReadIgnoresLegacyLastModified(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".metadata.json")
	if err := os.WriteFile(path, []byte(`{"type":"aws","lastModified":12345}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	manager := &MetadataManager{
		MetadataFilePath: path,
		ProviderDir:      dir,
		Metadata:         &CloudMetadata{},
	}
	if err := manager.Read(); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if manager.Metadata.Type != AWS {
		t.Fatalf("metadata type = %q, want %q", manager.Metadata.Type, AWS)
	}
	if manager.Metadata.Signature != "" {
		t.Fatalf("metadata signature = %q, want empty", manager.Metadata.Signature)
	}
}
