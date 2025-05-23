package gcp

import (
	"github.com/ip-api/cloudip/ip"
	// "cloudip/util" // No longer used
	// "fmt" // No longer used
)

// var appDir = util.GetAppDir() // Removed

const DataFile = "gcp.json" // Keep: specific to GCP
// const MetadataFile = ".metadata.json" // Removed: ip.GetMetadataFilePath uses ip.DefaultMetadataFile

func getDataUrl() string {
	return "https://www.gstatic.com/ipranges/cloud.json"
}

var ProviderDirectory = ip.GetProviderDirectory("gcp")
var DataFilePathGcp = ip.GetDataFilePath("gcp", DataFile)
var MetadataFilePathGcp = ip.GetMetadataFilePath("gcp")
