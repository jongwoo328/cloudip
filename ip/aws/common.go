package aws

import (
	"github.com/ip-api/cloudip/ip"
	// "cloudip/util" // No longer used
	// "fmt" // No longer used
)

// const DataFile = "aws.json" // This constant is used by DataFilePathAws below
const DataFile = "aws.json" // Keeping this as it's specific to AWS and used by ip.GetDataFilePath

// MetadataFile constant is removed as ip.GetMetadataFilePath uses ip.DefaultMetadataFile

func getDaraUrl() string {
	return "https://ip-ranges.amazonaws.com/ip-ranges.json"
}

var ProviderDirectory = ip.GetProviderDirectory("aws")
var DataFilePathAws = ip.GetDataFilePath("aws", DataFile) // DataFile is defined above
var MetadataFilePathAws = ip.GetMetadataFilePath("aws")
