package aws

import (
	"cloudip/internal"
	"fmt"
)

var appDir, _ = internal.GetAppDir()

const DataFile = "aws.json"
const MetadataFile = ".metadata.json"
const DataUrl = "https://ip-ranges.amazonaws.com/ip-ranges.json"

var DataFilePath = fmt.Sprintf("%s/%s/%s", appDir, "aws", DataFile)
var MetadataFilePath = fmt.Sprintf("%s/%s", appDir, MetadataFile)
