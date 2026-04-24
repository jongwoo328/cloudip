package cloudflare

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var appDir = util.GetAppDir(common.AppName)

const DataFileV4 = "cloudflare-v4.txt"
const DataFileV6 = "cloudflare-v6.txt"
const MetadataFile = ".metadata.json"

func getDataUrl() string {
	return "https://api.cloudflare.com/client/v4/ips"
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "cloudflare")
var DataFilePathCloudflareV4 = fmt.Sprintf("%s/%s", ProviderDirectory, DataFileV4)
var DataFilePathCloudflareV6 = fmt.Sprintf("%s/%s", ProviderDirectory, DataFileV6)
var MetadataFilePathCloudflare = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
