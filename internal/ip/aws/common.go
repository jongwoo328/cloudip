package aws

import (
	"cloudip/internal/util"
	"fmt"
)

var appDir, _ = util.GetAppDir()

const DataFile = "aws.json"
const MetadataFile = ".metadata.json"
const DataUrl = "https://ip-ranges.amazonaws.com/ip-ranges.json"

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "aws")
var DataFilePath = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePath = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
