package aws

import (
	"cloudip/util"
	"fmt"
)

var appDir, _ = util.GetAppDir()

const DataFile = "aws.json"
const MetadataFile = ".metadata.json"
const DataUrl = "https://ip-ranges.amazonaws.com/ip-ranges.json"

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "aws")
var DataFilePathAws = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathAws = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
