package gcp

import (
	"cloudip/common"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestGCPEnsureDataFileReusesFetchedDataForUpdate(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"syncToken": "new-sync-token",
			"creationTime": "2026-04-23T13:05:31.195904",
			"prefixes": []
		}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldMetadataManager := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &common.CloudMetadata{
			Type:      common.GCP,
			Signature: "old-sync-token",
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	manager := &IpDataManagerGcp{
		DataURI:      server.URL,
		DataFilePath: filepath.Join(dir, "gcp.json"),
	}
	if err := manager.writeData(&IpRangeDataGcp{SyncToken: "old-sync-token"}); err != nil {
		t.Fatalf("writeData() error = %v", err)
	}
	metadataManager.Metadata.Signature = "old-sync-token"
	requestCount = 0

	if err := manager.EnsureDataFile(); err != nil {
		t.Fatalf("EnsureDataFile() error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	if metadataManager.Metadata.Signature != "new-sync-token" {
		t.Fatalf("metadata signature = %q, want %q", metadataManager.Metadata.Signature, "new-sync-token")
	}
}
