package cloudflare

import (
	"cloudip/common"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCloudflareDownloadDataWritesCIDRsAndSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	v4Content, err := os.ReadFile(manager.DataFilePathV4)
	if err != nil {
		t.Fatalf("ReadFile(v4) error = %v", err)
	}
	if string(v4Content) != "173.245.48.0/20\n" {
		t.Fatalf("v4 content = %q, want %q", string(v4Content), "173.245.48.0/20\n")
	}

	v6Content, err := os.ReadFile(manager.DataFilePathV6)
	if err != nil {
		t.Fatalf("ReadFile(v6) error = %v", err)
	}
	if string(v6Content) != "2400:cb00::/32\n" {
		t.Fatalf("v6 content = %q, want %q", string(v6Content), "2400:cb00::/32\n")
	}

	if metadataManager.Metadata.Signature != "cf-etag" {
		t.Fatalf("metadata signature = %q, want %q", metadataManager.Metadata.Signature, "cf-etag")
	}

	data := manager.LoadIpData()
	if len(data.V4CIDRs) != 1 || data.V4CIDRs[0] != "173.245.48.0/20" {
		t.Fatalf("V4CIDRs = %#v, want Cloudflare IPv4 CIDR", data.V4CIDRs)
	}
	if len(data.V6CIDRs) != 1 || data.V6CIDRs[0] != "2400:cb00::/32" {
		t.Fatalf("V6CIDRs = %#v, want Cloudflare IPv6 CIDR", data.V6CIDRs)
	}
}
