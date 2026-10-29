package oci

import (
	"cloudip/common"
	"cloudip/ip"
	"cloudip/ip/provider"
	"cloudip/util"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A small representative snapshot of the official regions/cidrs/ipv6_cidrs schema.
const cachedData = `{"last_updated_timestamp":"old","regions":[{"region":"ap-seoul-1","cidrs":[{"cidr":"192.0.2.0/24","tags":["OCI"]},{"cidr":"198.51.100.0/24","tags":["OSN"]}],"ipv6_cidrs":[{"cidr":"2001:db8::/126","tags":["OCI"]}]},{"region":"us-phoenix-1","cidrs":[{"cidr":"203.0.113.0/24","tags":["OSN","OBJECT_STORAGE"]}]}]}`

var updatedData = strings.ReplaceAll(strings.ReplaceAll(cachedData, `"old"`, `"new"`), "192.0.2.0/24", "10.0.0.0/8")

func newTestManager(t *testing.T, cache string) *IpDataManagerOCI {
	t.Helper()
	dir := t.TempDir()
	original := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"), ProviderDir: dir,
		Metadata: &common.CloudMetadata{Type: common.OCI},
	}
	t.Cleanup(func() { metadataManager = original })
	if err := metadataManager.Ensure(); err != nil {
		t.Fatal(err)
	}
	if err := metadataManager.Write(&common.CloudMetadata{Type: common.OCI, Signature: "old", LastChecked: 1}); err != nil {
		t.Fatal(err)
	}
	manager := &IpDataManagerOCI{DataFilePath: filepath.Join(dir, DataFile)}
	if cache != "" {
		if err := os.WriteFile(manager.DataFilePath, []byte(cache), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return manager
}

func TestOCIValidateData(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"representative snapshot", cachedData, true},
		{"IPv4 only", `{"last_updated_timestamp":"now","regions":[{"region":"test","cidrs":[{"cidr":"192.0.2.0/24"}]}]}`, true},
		{"IPv6 only", `{"last_updated_timestamp":"now","regions":[{"region":"test","ipv6_cidrs":[{"cidr":"2001:db8::/32"}]}]}`, true},
		{"missing schema", `{}`, false},
		{"missing timestamp", strings.ReplaceAll(cachedData, `"old"`, `""`), false},
		{"empty ranges", `{"last_updated_timestamp":"now","regions":[{"region":"test","cidrs":[]}]}`, false},
		{"missing region", strings.ReplaceAll(cachedData, "ap-seoul-1", ""), false},
		{"bad IPv4", strings.ReplaceAll(cachedData, "192.0.2.0/24", "bad"), false},
		{"bad IPv6", strings.ReplaceAll(cachedData, "2001:db8::/126", "2001:db8::/129"), false},
		{"wrong family", strings.ReplaceAll(cachedData, "192.0.2.0/24", "2001:db8::/32"), false},
		{"mapped IPv6", strings.ReplaceAll(cachedData, "2001:db8::/126", "::ffff:192.0.2.0/120"), false},
		{"trailing JSON", cachedData + ` {}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var data IpRangeDataOCI
			err := util.DecodeJSON(strings.NewReader(tc.body), &data)
			if err == nil {
				err = data.validate()
			}
			if (err == nil) != tc.valid {
				t.Fatalf("valid = %v, error = %v", tc.valid, err)
			}
		})
	}
}

func TestOCIOfflineLazyInitializationAndMatching(t *testing.T) {
	manager := newTestManager(t, cachedData)
	p := newOCIProvider(manager)
	if _, err := p.CheckParsedIP(net.ParseIP("192.0.2.1")); err == nil {
		t.Fatal("constructor initialized the provider")
	}
	p.SetUpdatePolicy(common.UpdatePolicy{NoUpdate: true})
	checker := ip.NewIPChecker(map[common.CloudProvider]provider.CloudProvider{common.OCI: p}, []common.CloudProvider{common.OCI})
	for _, tc := range []struct {
		ip    string
		match bool
	}{
		{"192.0.2.0", true}, {"192.0.2.255", true}, {"192.0.3.0", false},
		{"198.51.100.1", true}, {"203.0.113.1", true},
		{"2001:db8::", true}, {"2001:db8::3", true}, {"2001:db8::4", false}, {"2001:db9::1", false},
	} {
		result := checker.Check([]string{tc.ip})[0]
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		if (result.Provider == common.OCI) != tc.match {
			t.Fatalf("%s: provider = %q", tc.ip, result.Provider)
		}
	}
	if _, err := p.CheckParsedIP(nil); err == nil {
		t.Fatal("nil IP accepted")
	}
	if err := os.Remove(manager.DataFilePath); err != nil {
		t.Fatal(err)
	}
	if err := p.Initialize(); err != nil {
		t.Fatalf("successful initialization not cached: %v", err)
	}
}

func TestOCIOfflineRetryAfterInitializationFailure(t *testing.T) {
	manager := newTestManager(t, "")
	p := newOCIProvider(manager)
	p.SetUpdatePolicy(common.UpdatePolicy{NoUpdate: true})
	if err := p.Initialize(); err == nil || !strings.Contains(err.Error(), "--no-update") {
		t.Fatalf("missing cache: %v", err)
	}
	if err := os.WriteFile(manager.DataFilePath, []byte(cachedData), 0644); err != nil {
		t.Fatal(err)
	}
	if err := p.Initialize(); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	match, err := p.CheckParsedIP(net.ParseIP("192.0.2.1"))
	if err != nil || !match {
		t.Fatalf("retry lost ranges: %v %v", match, err)
	}
}

func TestOCIOfflineRejectsInvalidCache(t *testing.T) {
	for _, cache := range []string{"{}", strings.ReplaceAll(cachedData, "192.0.2.0/24", "bad"), cachedData + ` {}`} {
		t.Run(cache[:2], func(t *testing.T) {
			manager := newTestManager(t, cache)
			manager.SetUpdatePolicy(common.UpdatePolicy{NoUpdate: true})
			if err := newOCIProvider(manager).Initialize(); err == nil {
				t.Fatal("invalid cache accepted")
			}
			failure := errors.New("update failed")
			if err := manager.useCacheIfUpdateFails(failure); !errors.Is(err, failure) {
				t.Fatalf("update error hidden: %v", err)
			}
		})
	}
}

func TestOCIOfflineFreshCacheSkipsNetwork(t *testing.T) {
	manager := newTestManager(t, cachedData)
	// An invalid URL makes accidental network access fail immediately.
	manager.DataURI = ":invalid:"
	if err := metadataManager.MarkChecked(time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := manager.EnsureDataFile(); err != nil {
		t.Fatal(err)
	}
	if metadataManager.Metadata.Signature != "old" {
		t.Fatal("fresh signature changed")
	}
}

func TestOCIUpdateFailuresPreserveCache(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		truncated  bool
	}{
		{"malformed JSON", "{", 200, false},
		{"empty schema", "{}", 200, false},
		{"empty ranges", `{"last_updated_timestamp":"new","regions":[]}`, 200, false},
		{"invalid IPv4", strings.ReplaceAll(updatedData, "10.0.0.0/8", "bad"), 200, false},
		{"invalid IPv6", strings.ReplaceAll(updatedData, "2001:db8::/126", "bad"), 200, false},
		{"duplicate JSON names", strings.TrimSuffix(updatedData, "}") + `,"ignored":1,"ignored":2}`, 200, false},
		{"invalid UTF-8", strings.TrimSuffix(updatedData, "}") + ",\"ignored\":\"\xff\"}", 200, false},
		{"trailing JSON", updatedData + ` {}`, 200, false},
		{"HTTP failure", "unavailable", 503, false},
		{"truncated response", updatedData, 200, true},
	} {
		for _, withCache := range []bool{true, false} {
			label := "initial/"
			if withCache {
				label = "update/"
			}
			t.Run(label+tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.truncated {
						w.Header().Set("Content-Length", "10000")
					}
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.body))
				}))
				defer server.Close()
				cache := ""
				if withCache {
					cache = cachedData
				}
				manager := newTestManager(t, cache)
				manager.DataURI = server.URL
				err := manager.EnsureDataFile()
				if withCache {
					if err != nil {
						t.Fatal(err)
					}
					b, err := os.ReadFile(manager.DataFilePath)
					if err != nil {
						t.Fatal(err)
					}
					if string(b) != cachedData {
						t.Fatal("valid cache replaced")
					}
					if _, err := manager.LoadIpData(); err != nil {
						t.Fatal(err)
					}
				} else {
					if err == nil {
						t.Fatal("first download accepted invalid data")
					}
					if _, err := os.Stat(manager.DataFilePath); !os.IsNotExist(err) {
						t.Fatalf("invalid data published: %v", err)
					}
				}
				if metadataManager.Metadata.Signature != "old" || metadataManager.Metadata.LastChecked != 1 {
					t.Fatalf("metadata changed: %+v", metadataManager.Metadata)
				}
			})
		}
	}
}

func TestOCIUpdatePublishesOnceAndSkipsFreshChecks(t *testing.T) {
	for _, existing := range []bool{false, true} {
		name := "initial"
		if existing {
			name = "refresh"
		}
		t.Run(name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++; _, _ = w.Write([]byte(updatedData)) }))
			defer server.Close()
			cache := ""
			if existing {
				cache = cachedData
			}
			manager := newTestManager(t, cache)
			manager.DataURI = server.URL
			if existing {
				if _, err := manager.LoadIpData(); err != nil {
					t.Fatal(err)
				}
			}
			if err := manager.EnsureDataFile(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatalf("requests = %d", requests)
			}
			if metadataManager.Metadata.Signature != "new" || metadataManager.Metadata.LastChecked <= 1 {
				t.Fatalf("metadata not refreshed: %+v", metadataManager.Metadata)
			}
			data, err := manager.LoadIpData()
			if err != nil {
				t.Fatal(err)
			}
			if data.LastUpdatedTimestamp != "new" {
				t.Fatal("in-memory cache is stale")
			}
			if err := manager.EnsureDataFile(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatalf("fresh cache made %d requests", requests)
			}
			if err := metadataManager.Write(&common.CloudMetadata{Type: common.OCI, Signature: "new", LastChecked: 1}); err != nil {
				t.Fatal(err)
			}
			if err := manager.EnsureDataFile(); err != nil {
				t.Fatal(err)
			}
			if requests != 2 || metadataManager.Metadata.LastChecked <= 1 {
				t.Fatal("unchanged response not marked checked")
			}
		})
	}
}
