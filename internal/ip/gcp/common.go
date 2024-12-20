package gcp

import (
	"cloudip/internal/util"
	"fmt"
)

var appDir, _ = util.GetAppDir()

const DataFile = "gcp.json"
const MetadataFile = ".metadata.json"
const DataUrl = "https://www.gstatic.com/ipranges/cloud.json"

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "gcp")
var DataFilePathAws = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathAws = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
