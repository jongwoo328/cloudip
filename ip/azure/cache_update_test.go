package azure

import (
	"cloudip/common"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAzureUpdateRejectsBadDataWithoutLosingCache(t *testing.T) {
	const valid = `{"changeNumber":2,"values":[{"properties":{"addressPrefixes":["10.0.0.0/8","2001:db8::/32"]}}]}`
	const cached = `{"changeNumber":2,"values":[{"properties":{"addressPrefixes":["192.0.2.0/24","2001:db8::/32"]}}]}`
	for _, tc := range []struct {
		name, body string
		truncated  bool
	}{
		{"malformed JSON", "{", false},
		{"missing schema", "{}", false},
		{"empty ranges", `{"changeNumber":2,"values":[]}`, false},
		{"invalid CIDR", strings.ReplaceAll(valid, "10.0.0.0/8", "bad-cidr"), false},
		{"trailing JSON", valid + " {}", false},
		{"truncated response", valid, true},
	} {
		for _, withCache := range []bool{true, false} {
			label := "initial/"
			if withCache {
				label = "update/"
			}
			t.Run(label+tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Last-Modified", "Mon, 01 Jun 2026 00:00:00 GMT")
					w.Header().Set("ETag", `"new"`)
					if tc.truncated {
						w.Header().Set("Content-Length", "10000")
					}
					if r.Method != http.MethodHead {
						w.Write([]byte(tc.body))
					}
				}))
				defer server.Close()
				dir := t.TempDir()
				original := metadataManager
				metadataManager = &common.MetadataManager{MetadataFilePath: filepath.Join(dir, ".metadata.json"), ProviderDir: dir, Metadata: &common.CloudMetadata{Type: common.Azure}}
				t.Cleanup(func() { metadataManager = original })
				if err := metadataManager.Write(&common.CloudMetadata{Type: common.Azure, Signature: "old", LastChecked: 1}); err != nil {
					t.Fatal(err)
				}
				manager := &IpDataManagerAzure{DataURI: server.URL, DataFilePath: filepath.Join(dir, "azure.json")}
				path := manager.DataFilePath
				if withCache {
					if err := os.WriteFile(path, []byte(cached), 0644); err != nil {
						t.Fatal(err)
					}
				}
				err := manager.EnsureDataFile()
				if withCache {
					if err != nil {
						t.Fatalf("valid cache should keep lookup available: %v", err)
					}
					contents, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if string(contents) != cached {
						t.Fatalf("cache was changed: %s", contents)
					}
					if _, err := manager.LoadIpData(); err != nil {
						t.Fatalf("preserved cache cannot be loaded: %v", err)
					}
				} else {
					if err == nil {
						t.Fatal("first download must fail without a valid cache")
					}
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("bad cache was published: %v", err)
					}
				}
				if metadataManager.Metadata.Signature != "old" || metadataManager.Metadata.LastChecked != 1 {
					t.Fatalf("metadata changed after failed update: %+v", metadataManager.Metadata)
				}
			})
		}
	}
}

func TestAzureValidUpdatePublishesCacheAndMetadata(t *testing.T) {
	const valid = `{"changeNumber":2,"values":[{"properties":{"addressPrefixes":["10.0.0.0/8","2001:db8::/32"]}}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Last-Modified", "Mon, 01 Jun 2026 00:00:00 GMT")
		w.Header().Set("ETag", `"new"`)
		if r.Method != http.MethodHead {
			w.Write([]byte(valid))
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	original := metadataManager
	metadataManager = &common.MetadataManager{MetadataFilePath: filepath.Join(dir, ".metadata.json"), ProviderDir: dir, Metadata: &common.CloudMetadata{Type: common.Azure}}
	t.Cleanup(func() { metadataManager = original })
	if err := metadataManager.Write(&common.CloudMetadata{Type: common.Azure, Signature: "old", LastChecked: 1}); err != nil {
		t.Fatal(err)
	}
	manager := &IpDataManagerAzure{DataURI: server.URL, DataFilePath: filepath.Join(dir, "azure.json")}
	if err := os.WriteFile(manager.DataFilePath, []byte(`{"changeNumber":2,"values":[{"properties":{"addressPrefixes":["192.0.2.0/24","2001:db8::/32"]}}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := manager.EnsureDataFile(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.LoadIpData(); err != nil {
		t.Fatal(err)
	}
	if metadataManager.Metadata.Signature == "old" || metadataManager.Metadata.LastChecked <= 1 {
		t.Fatalf("metadata not updated: %+v", metadataManager.Metadata)
	}
}

func TestAzureInvalidCacheDoesNotHideUpdateError(t *testing.T) {
	dir := t.TempDir()
	manager := &IpDataManagerAzure{DataFilePath: filepath.Join(dir, "data.json")}
	if err := os.WriteFile(manager.DataFilePath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("update failed")
	if err := manager.useCacheIfUpdateFails(failure); !errors.Is(err, failure) {
		t.Fatalf("update failure hidden by invalid cache: %v", err)
	}
}
