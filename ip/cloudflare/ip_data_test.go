package cloudflare

import (
	"cloudip/common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCloudflareDownloadDataWritesCIDRsAndSignature(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"result": {
				"etag": "cf-etag",
				"ipv4_cidrs": ["173.245.48.0/20"],
				"ipv6_cidrs": ["2400:cb00::/32"]
			}
		}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldMetadataManager := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &common.CloudMetadata{
			Type: common.Cloudflare,
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	manager := &IpDataManagerCloudflare{
		DataURI:        server.URL,
		DataFilePathV4: filepath.Join(dir, "cloudflare-v4.txt"),
		DataFilePathV6: filepath.Join(dir, "cloudflare-v6.txt"),
	}

	if err := manager.EnsureDataFile(); err != nil {
		t.Fatalf("EnsureDataFile() error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}

	content, err := os.ReadFile(manager.snapshotPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "173.245.48.0/20") || !strings.Contains(string(content), "2400:cb00::/32") {
		t.Fatalf("snapshot missing address families: %s", content)
	}

	if metadataManager.Metadata.Signature != "cf-etag" {
		t.Fatalf("metadata signature = %q, want %q", metadataManager.Metadata.Signature, "cf-etag")
	}

	data, err := manager.LoadIpData()
	if err != nil {
		t.Fatalf("LoadIpData() error = %v", err)
	}
	if len(data.V4CIDRs) != 1 || data.V4CIDRs[0] != "173.245.48.0/20" {
		t.Fatalf("V4CIDRs = %#v, want Cloudflare IPv4 CIDR", data.V4CIDRs)
	}
	if len(data.V6CIDRs) != 1 || data.V6CIDRs[0] != "2400:cb00::/32" {
		t.Fatalf("V6CIDRs = %#v, want Cloudflare IPv6 CIDR", data.V6CIDRs)
	}
}

func TestCloudflareEnsureDataFileReusesFetchedDataForUpdate(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"result": {
				"etag": "new-etag",
				"ipv4_cidrs": ["173.245.48.0/20"],
				"ipv6_cidrs": ["2400:cb00::/32"]
			}
		}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldMetadataManager := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &common.CloudMetadata{
			Type:      common.Cloudflare,
			Signature: "old-etag",
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	manager := &IpDataManagerCloudflare{
		DataURI:        server.URL,
		DataFilePathV4: filepath.Join(dir, "cloudflare-v4.txt"),
		DataFilePathV6: filepath.Join(dir, "cloudflare-v6.txt"),
	}
	if err := writeCIDRLines(manager.DataFilePathV4, []string{"10.0.0.0/8"}); err != nil {
		t.Fatalf("writeCIDRLines(v4) error = %v", err)
	}
	if err := writeCIDRLines(manager.DataFilePathV6, []string{"2001:db8::/32"}); err != nil {
		t.Fatalf("writeCIDRLines(v6) error = %v", err)
	}

	if err := manager.EnsureDataFile(); err != nil {
		t.Fatalf("EnsureDataFile() error = %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	if metadataManager.Metadata.Signature != "new-etag" {
		t.Fatalf("metadata signature = %q, want %q", metadataManager.Metadata.Signature, "new-etag")
	}
	data, err := manager.LoadIpData()
	if err != nil {
		t.Fatal(err)
	}
	if len(data.V4CIDRs) != 1 || data.V4CIDRs[0] != "173.245.48.0/20" || len(data.V6CIDRs) != 1 || data.V6CIDRs[0] != "2400:cb00::/32" {
		t.Fatalf("legacy cache was not replaced by a complete snapshot: %+v", data)
	}
}

func TestCloudflareLoadIpDataReturnsErrorForMissingFile(t *testing.T) {
	dir := t.TempDir()
	manager := &IpDataManagerCloudflare{
		DataFilePathV4: filepath.Join(dir, "missing-v4.txt"),
		DataFilePathV6: filepath.Join(dir, "missing-v6.txt"),
	}

	if _, err := manager.LoadIpData(); err == nil {
		t.Fatal("LoadIpData() error = nil, want error")
	}
}

func TestCloudflareEnsureDataFileSkipsFreshUpdateCheck(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"result": {
				"etag": "new-etag",
				"ipv4_cidrs": ["173.245.48.0/20"],
				"ipv6_cidrs": ["2400:cb00::/32"]
			}
		}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldMetadataManager := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &common.CloudMetadata{
			Type: common.Cloudflare,
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	if err := metadataManager.Write(&common.CloudMetadata{
		Type:        common.Cloudflare,
		Signature:   "old-etag",
		LastChecked: time.Now().Unix(),
	}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	v4Path := filepath.Join(dir, "cloudflare-v4.txt")
	v6Path := filepath.Join(dir, "cloudflare-v6.txt")
	if err := writeCIDRLines(v4Path, []string{"10.0.0.0/8"}); err != nil {
		t.Fatalf("writeCIDRLines(v4) error = %v", err)
	}
	if err := writeCIDRLines(v6Path, []string{"2001:db8::/32"}); err != nil {
		t.Fatalf("writeCIDRLines(v6) error = %v", err)
	}

	manager := &IpDataManagerCloudflare{
		DataURI:        server.URL,
		DataFilePathV4: v4Path,
		DataFilePathV6: v6Path,
	}

	if err := manager.EnsureDataFile(); err != nil {
		t.Fatalf("EnsureDataFile() error = %v", err)
	}
	if requestCount != 0 {
		t.Fatalf("request count = %d, want 0", requestCount)
	}
	if metadataManager.Metadata.Signature != "old-etag" {
		t.Fatalf("metadata signature = %q, want unchanged signature", metadataManager.Metadata.Signature)
	}
}

func TestCloudflareEnsureDataFileNoUpdateRequiresExistingFiles(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"success": true,
			"result": {
				"etag": "etag",
				"ipv4_cidrs": [],
				"ipv6_cidrs": []
			}
		}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	oldMetadataManager := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"),
		ProviderDir:      dir,
		Metadata: &common.CloudMetadata{
			Type: common.Cloudflare,
		},
	}
	t.Cleanup(func() {
		metadataManager = oldMetadataManager
	})

	manager := &IpDataManagerCloudflare{
		DataURI:        server.URL,
		DataFilePathV4: filepath.Join(dir, "missing-v4.txt"),
		DataFilePathV6: filepath.Join(dir, "missing-v6.txt"),
		UpdatePolicy:   common.UpdatePolicy{NoUpdate: true},
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

func writeCIDRLines(path string, cidrs []string) error {
	return os.WriteFile(path, []byte(strings.Join(cidrs, "\n")+"\n"), 0644)
}

func TestCloudflareSnapshotAndLegacyCachesWorkOffline(t *testing.T) {
	for _, snapshot := range []bool{false, true} {
		for _, noUpdate := range []bool{false, true} {
			t.Run(fmt.Sprintf("snapshot=%t/no-update=%t", snapshot, noUpdate), func(t *testing.T) {
				requestCount := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestCount++
					w.WriteHeader(http.StatusServiceUnavailable)
				}))
				defer server.Close()
				dir := t.TempDir()
				original := metadataManager
				metadataManager = &common.MetadataManager{MetadataFilePath: filepath.Join(dir, ".metadata.json"), ProviderDir: dir, Metadata: &common.CloudMetadata{Type: common.Cloudflare}}
				t.Cleanup(func() { metadataManager = original })
				if err := metadataManager.Write(&common.CloudMetadata{Type: common.Cloudflare, Signature: "cached", LastChecked: time.Now().Unix()}); err != nil {
					t.Fatal(err)
				}
				manager := &IpDataManagerCloudflare{DataURI: server.URL, DataFilePathV4: filepath.Join(dir, "v4.txt"), DataFilePathV6: filepath.Join(dir, "v6.txt"), UpdatePolicy: common.UpdatePolicy{NoUpdate: noUpdate}}
				if snapshot {
					if err := os.WriteFile(manager.snapshotPath(), []byte(`{"ipv4_cidrs":["10.0.0.0/8"],"ipv6_cidrs":["2001:db8::/32"]}`), 0644); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := writeCIDRLines(manager.DataFilePathV4, []string{"10.0.0.0/8"}); err != nil {
						t.Fatal(err)
					}
					if err := writeCIDRLines(manager.DataFilePathV6, []string{"2001:db8::/32"}); err != nil {
						t.Fatal(err)
					}
				}
				if err := manager.EnsureDataFile(); err != nil {
					t.Fatal(err)
				}
				if requestCount != 0 {
					t.Fatalf("offline cache made %d requests", requestCount)
				}
				data, err := manager.LoadIpData()
				if err != nil {
					t.Fatal(err)
				}
				if len(data.V4CIDRs) != 1 || len(data.V6CIDRs) != 1 {
					t.Fatalf("missing cached address family: %+v", data)
				}
			})
		}
	}
}
