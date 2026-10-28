package digitalocean

import (
	"cloudip/common"
	"cloudip/ip"
	"cloudip/ip/provider"
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

const cachedData = `{"cidrs":["192.0.2.0/24","198.51.100.0/24","2001:db8::/126","203.0.113.0/24"]}`
const updatedData = "10.0.0.0/8,KR,KR-11,Seoul,\n198.51.100.0/24,US,US-AZ,Phoenix,85001\n2001:db8::/126,KR,KR-11,Seoul,\n203.0.113.0/24,US,US-AZ,Phoenix,85001\n"

func newTestManager(t *testing.T, cache string) *IpDataManagerDigitalOcean {
	t.Helper()
	dir := t.TempDir()
	original := metadataManager
	metadataManager = &common.MetadataManager{
		MetadataFilePath: filepath.Join(dir, ".metadata.json"), ProviderDir: dir,
		Metadata: &common.CloudMetadata{Type: common.DigitalOcean},
	}
	t.Cleanup(func() { metadataManager = original })
	if err := metadataManager.Ensure(); err != nil {
		t.Fatal(err)
	}
	if err := metadataManager.Write(&common.CloudMetadata{Type: common.DigitalOcean, Signature: "old", LastChecked: 1}); err != nil {
		t.Fatal(err)
	}
	manager := &IpDataManagerDigitalOcean{DataFilePath: filepath.Join(dir, DataFile)}
	if cache != "" {
		if err := os.WriteFile(manager.DataFilePath, []byte(cache), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return manager
}

func TestDigitalOceanValidateData(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"representative snapshot", updatedData, true},
		{"IPv4 only", "192.0.2.0/24,KR,KR-11,Seoul,\n", true},
		{"IPv6 only", "2001:db8::/32,KR,KR-11,Seoul,\n", true},
		{"CRLF and blank lines", "\r\n" + strings.ReplaceAll(updatedData, "\n", "\r\n"), true},
		{"comments", "# feed, with, commas\n" + updatedData, true},
		{"quoted city", "192.0.2.0/24,KR,KR-11,\"Seoul, City\",\n", true},
		{"empty geographic fields", "192.0.2.0/24,,,,\n", true},
		{"empty feed", "", false},
		{"only comments", "# nothing here\n", false},
		{"HTML", "<html>error</html>", false},
		{"missing columns", "192.0.2.0/24,KR\n", false},
		{"extra columns", "192.0.2.0/24,KR,KR-11,Seoul,,extra\n", false},
		{"malformed CSV", "192.0.2.0/24,KR,KR-11,\"Seoul,\n", false},
		{"bad IPv4", strings.ReplaceAll(updatedData, "10.0.0.0/8", "bad"), false},
		{"bad IPv6", strings.ReplaceAll(updatedData, "2001:db8::/126", "2001:db8::/129"), false},
		{"mapped IPv6", strings.ReplaceAll(updatedData, "2001:db8::/126", "::ffff:192.0.2.0/120"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeGeofeed(strings.NewReader(tc.body))
			if (err == nil) != tc.valid {
				t.Fatalf("valid = %v, error = %v", tc.valid, err)
			}
		})
	}
}

func TestDigitalOceanValidateSignatureTracksRanges(t *testing.T) {
	a, err := decodeGeofeed(strings.NewReader(updatedData))
	if err != nil {
		t.Fatal(err)
	}
	b, err := decodeGeofeed(strings.NewReader("# changed header\n" + strings.ReplaceAll(updatedData, "Seoul", "Other city")))
	if err != nil {
		t.Fatal(err)
	}
	if a.signature() != b.signature() {
		t.Fatal("geographic changes altered IP range signature")
	}
	b.CIDRs[0] = "192.0.2.0/24"
	if a.signature() == b.signature() {
		t.Fatal("changed IP ranges have identical signatures")
	}
}

func TestDigitalOceanOfflineLazyInitializationAndMatching(t *testing.T) {
	manager := newTestManager(t, cachedData)
	p := newDigitalOceanProvider(manager)
	if _, err := p.CheckParsedIP(net.ParseIP("192.0.2.1")); err == nil {
		t.Fatal("constructor initialized the provider")
	}
	p.SetUpdatePolicy(common.UpdatePolicy{NoUpdate: true})
	checker := ip.NewIPChecker(map[common.CloudProvider]provider.CloudProvider{common.DigitalOcean: p}, []common.CloudProvider{common.DigitalOcean})
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
		if (result.Provider == common.DigitalOcean) != tc.match {
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

func TestDigitalOceanOfflineRetryAfterInitializationFailure(t *testing.T) {
	manager := newTestManager(t, "")
	p := newDigitalOceanProvider(manager)
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

func TestDigitalOceanOfflineRejectsInvalidCache(t *testing.T) {
	for _, cache := range []string{"{}", strings.ReplaceAll(cachedData, "192.0.2.0/24", "bad"), cachedData + ` {}`} {
		t.Run(cache[:2], func(t *testing.T) {
			manager := newTestManager(t, cache)
			manager.SetUpdatePolicy(common.UpdatePolicy{NoUpdate: true})
			if err := newDigitalOceanProvider(manager).Initialize(); err == nil {
				t.Fatal("invalid cache accepted")
			}
			failure := errors.New("update failed")
			if err := manager.useCacheIfUpdateFails(failure); !errors.Is(err, failure) {
				t.Fatalf("update error hidden: %v", err)
			}
		})
	}
}

func TestDigitalOceanOfflineFreshCacheSkipsNetwork(t *testing.T) {
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

func TestDigitalOceanUpdateFailuresPreserveCache(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		truncated  bool
	}{
		{"malformed CSV", `192.0.2.0/24,KR,KR-11,"Seoul,`, 200, false},
		{"unexpected JSON", "{}", 200, false},
		{"empty ranges", "# no ranges\n", 200, false},
		{"invalid IPv4", strings.ReplaceAll(updatedData, "10.0.0.0/8", "bad"), 200, false},
		{"invalid IPv6", strings.ReplaceAll(updatedData, "2001:db8::/126", "bad"), 200, false},
		{"trailing bad CSV", updatedData + `bad`, 200, false},
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

func TestDigitalOceanUpdatePublishesOnceAndSkipsFreshChecks(t *testing.T) {
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
			if metadataManager.Metadata.Signature == "old" || metadataManager.Metadata.Signature == "" || metadataManager.Metadata.LastChecked <= 1 {
				t.Fatalf("metadata not refreshed: %+v", metadataManager.Metadata)
			}
			data, err := manager.LoadIpData()
			if err != nil {
				t.Fatal(err)
			}
			if data.CIDRs[0] != "10.0.0.0/8" {
				t.Fatal("in-memory cache is stale")
			}
			if err := manager.EnsureDataFile(); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatalf("fresh cache made %d requests", requests)
			}
			if err := metadataManager.Write(&common.CloudMetadata{Type: common.DigitalOcean, Signature: metadataManager.Metadata.Signature, LastChecked: 1}); err != nil {
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
