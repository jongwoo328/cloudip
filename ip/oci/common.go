package oci

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var appDir = util.GetAppDir(common.AppName)

const DataFile = "oci.json"
const MetadataFile = ".metadata.json"

func getDataUrl() string {
	return "https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json"
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "oci")
var DataFilePathOCI = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathOCI = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
