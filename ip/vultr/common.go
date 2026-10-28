package vultr

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var appDir = util.GetAppDir(common.AppName)

const DataFile = "vultr.json"
const MetadataFile = ".metadata.json"

func getDataUrl() string {
	return "https://geofeed.constant.com/?json"
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "vultr")
var DataFilePathVultr = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathVultr = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
