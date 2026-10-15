package common

import (
	"errors"
	"syscall"
	"testing"
)

func TestMetadataEnsureReturnsStatError(t *testing.T) {
	m := &MetadataManager{MetadataFilePath: "bad\x00path", ProviderDir: t.TempDir(), Metadata: &CloudMetadata{Type: AWS}}
	if err := m.Ensure(); !errors.Is(err, syscall.EINVAL) {
		t.Fatalf("Ensure error=%v, want original stat error", err)
	}
}
