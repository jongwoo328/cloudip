package gcp

import (
	"cloudip/common"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	if err := metadataManager.Write(&common.CloudMetadata{
		Type:        common.GCP,
		Signature:   "old-sync-token",
		LastChecked: time.Now().Add(-48 * time.Hour).Unix(),
	}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
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

func TestGCPLoadIpDataReturnsErrorForMissingFile(t *testing.T) {
	dir := t.TempDir()
	manager := &IpDataManagerGcp{
		DataFilePath: filepath.Join(dir, "missing.json"),
	}

	if _, err := manager.LoadIpData(); err == nil {
		t.Fatal("LoadIpData() error = nil, want error")
	}
}

func TestGCPEnsureDataFileSkipsFreshUpdateCheck(t *testing.T) {
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
			Type: common.GCP,
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	if err := metadataManager.Write(&common.CloudMetadata{
		Type:        common.GCP,
		Signature:   "old-sync-token",
		LastChecked: time.Now().Unix(),
	}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	dataPath := filepath.Join(dir, "gcp.json")
	if err := os.WriteFile(dataPath, []byte(`{"syncToken":"old-sync-token","prefixes":[]}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	manager := &IpDataManagerGcp{
		DataURI:      server.URL,
		DataFilePath: dataPath,
	}

	if err := manager.EnsureDataFile(); err != nil {
		t.Fatalf("EnsureDataFile() error = %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("request count = %d, want 0", requestCount)
	}
	if metadataManager.Metadata.Signature != "old-sync-token" {
		t.Fatalf("metadata signature = %q, want unchanged signature", metadataManager.Metadata.Signature)
	}
}

func TestGCPEnsureDataFileNoUpdateRequiresExistingFile(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"syncToken":"sync-token","prefixes":[]}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldMetadataManager := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &common.CloudMetadata{
			Type: common.GCP,
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	manager := &IpDataManagerGcp{
		DataURI:      server.URL,
		DataFilePath: filepath.Join(dir, "missing.json"),
		UpdatePolicy: common.UpdatePolicy{NoUpdate: true},
	}

	err := manager.EnsureDataFile()
	if err == nil {
		t.Fatal("EnsureDataFile() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "--no-update") {
		t.Fatalf("error = %v, want --no-update context", err)
	}
	if requestCount != 0 {
		t.Fatalf("request count = %d, want 0", requestCount)
	}
}
