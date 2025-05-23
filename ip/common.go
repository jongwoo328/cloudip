package ip

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
)

var AppDir = util.GetAppDir()

const DefaultMetadataFile = ".metadata.json"

// GetProviderDirectory returns the path to the provider-specific directory.
func GetProviderDirectory(providerName string) string {
	return fmt.Sprintf("%s/%s", AppDir, providerName)
}

// GetDataFilePath returns the full path to a provider's data file.
func GetDataFilePath(providerName string, dataFileName string) string {
	return fmt.Sprintf("%s/%s", GetProviderDirectory(providerName), dataFileName)
}

// GetMetadataFilePath returns the full path to a provider's metadata file.
func GetMetadataFilePath(providerName string) string {
	return fmt.Sprintf("%s/%s", GetProviderDirectory(providerName), DefaultMetadataFile)
}

var Providers = []common.CloudProvider{
	"aws",
	"gcp",
	"azure",
}
