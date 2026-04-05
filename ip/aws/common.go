package aws

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var appDir = util.GetAppDir(common.AppName)

const DataFile = "aws.json"
const MetadataFile = ".metadata.json"

func getDataUrl() string {
	return "https://ip-ranges.amazonaws.com/ip-ranges.json"
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "aws")
var DataFilePathAws = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathAws = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
