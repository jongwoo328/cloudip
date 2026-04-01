package gcp

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var appDir = util.GetAppDir(common.AppName)

const DataFile = "gcp.json"
const MetadataFile = ".metadata.json"

func getDataUrl() string {
	return "https://www.gstatic.com/ipranges/cloud.json"
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "gcp")
var DataFilePathGcp = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathGcp = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
