package digitalocean

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var appDir = util.GetAppDir(common.AppName)

const DataFile = "digitalocean.json"
const MetadataFile = ".metadata.json"

func getDataUrl() string {
	return "https://www.digitalocean.com/geo/google.csv"
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "digitalocean")
var DataFilePathDigitalOcean = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathDigitalOcean = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
