package util_test

import (
	"bytes"
	"cloudip/common"
	"cloudip/ip/aws"
	"cloudip/ip/azure"
	"cloudip/ip/cloudflare"
	"cloudip/ip/digitalocean"
	"cloudip/ip/gcp"
	"cloudip/ip/oci"
	"cloudip/ip/vultr"
	"cloudip/util"
	legacyjson "encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The legacy API is deliberately retained as an independent format oracle.
func checkLegacyCache[T any](t *testing.T, input string) {
	t.Helper()
	var original T
	if err := legacyjson.Unmarshal([]byte(input), &original); err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	encoder := legacyjson.NewEncoder(&expected)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(&original); err != nil {
		t.Fatal(err)
	}
	// omitempty can omit an empty slice, so compare both readers on the same
	// persisted bytes rather than the original in-memory nil/empty distinction.
	var legacyRestored T
	if err := legacyjson.Unmarshal(expected.Bytes(), &legacyRestored); err != nil {
		t.Fatal(err)
	}

	var restored T
	if err := util.DecodeJSON(bytes.NewReader(expected.Bytes()), &restored); err != nil {
		t.Fatalf("legacy cache cannot be read: %v", err)
	}
	if !reflect.DeepEqual(legacyRestored, restored) {
		t.Fatalf("legacy cache changed decoded values for %T", restored)
	}
	path := filepath.Join(t.TempDir(), "cache.json")
	if err := util.WriteJSONFile(path, &restored); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected.Bytes()) {
		t.Fatalf("cache format changed:\ngot  %s\nwant %s", actual, expected.Bytes())
	}
}

func TestJSONLegacyCacheCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		check       func(*testing.T, string)
	}{
		{"AWS", `{"syncToken":"old","prefixes":[{"ip_prefix":"192.0.2.0/24"}],"ipv6_prefixes":null}`, checkLegacyCache[aws.IpRangeDataAws]},
		{"GCP", `{"syncToken":"old","prefixes":[{"ipv4Prefix":"192.0.2.0/24"}]}`, checkLegacyCache[gcp.IpRangeDataGcp]},
		{"Azure", `{"changeNumber":1,"values":[{"properties":{"addressPrefixes":["192.0.2.0/24"],"networkFeatures":null}}]}`, checkLegacyCache[azure.IpRangeDataAzure]},
		{"Cloudflare", `{"ipv4_cidrs":["192.0.2.0/24"],"ipv6_cidrs":null}`, checkLegacyCache[cloudflare.IpRangeDataCloudflare]},
		{"OCI", `{"last_updated_timestamp":"old","regions":[{"region":"test","cidrs":[{"cidr":"192.0.2.0/24"}]}]}`, checkLegacyCache[oci.IpRangeDataOCI]},
		{"OCI empty optional arrays", `{"last_updated_timestamp":"old","regions":[{"region":"test","cidrs":[{"cidr":"192.0.2.0/24","tags":[]}],"ipv6_cidrs":[]}]}`, checkLegacyCache[oci.IpRangeDataOCI]},
		{"Vultr", `{"updated":"old","subnets":[{"ip_prefix":"192.0.2.0/24"}]}`, checkLegacyCache[vultr.IpRangeDataVultr]},
		{"DigitalOcean", `{"cidrs":["192.0.2.0/24"]}`, checkLegacyCache[digitalocean.IpRangeDataDigitalOcean]},
		{"metadata unchecked", `{"type":"aws","signature":"old"}`, checkLegacyCache[common.CloudMetadata]},
		{"metadata checked", `{"type":"aws","signature":"old","lastChecked":42}`, checkLegacyCache[common.CloudMetadata]},
	} {
		t.Run(tc.name, func(t *testing.T) { tc.check(t, tc.input) })
	}
}

func TestJSONCacheFormattingCompatibility(t *testing.T) {
	type cache struct {
		Names    map[string]string `json:"names"`
		NilMap   map[string]string `json:"nilMap"`
		NilSlice []string          `json:"nilSlice"`
		Empty    []string          `json:"empty"`
	}
	checkLegacyCache[cache](t, `{"names":{"z":"<&>\u2028\u2029","a":"한글"},"nilMap":null,"nilSlice":null,"empty":[]}`)
}
