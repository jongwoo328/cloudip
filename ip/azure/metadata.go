package azure

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
	// "io" // No longer used
	// "os" // No longer used
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAzure,
	Metadata: &common.CloudMetadata{
		Type:         common.Azure,
		LastModified: 0,
	},
	// Add a SaveFunc that uses the generic WriteMetadata
	SaveFunc: func(filePath string, data *common.CloudMetadata) error {
		return common.WriteMetadata(filePath, data)
	},
}

func init() {
	initialMeta := &common.CloudMetadata{Type: common.Azure, LastModified: 0}
	err := common.EnsureMetadataFile(MetadataFilePathAzure, ProviderDirectory, common.Azure, initialMeta)
	if err != nil {
		util.PrintErrorTrace(fmt.Errorf("Error ensuring Azure metadata file during init: %w", err))
		return
	}

	err = common.ReadMetadata(MetadataFilePathAzure, metadataManager.Metadata)
	if err != nil {
		util.PrintErrorTrace(fmt.Errorf("Error reading Azure metadata during init: %w", err))
		return
	}
}

// ensureMetadataFile, readMetadata, and writeMetadata are removed,
// their functionality is replaced by common.EnsureMetadataFile,
// common.ReadMetadata, and common.WriteMetadata respectively.

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAzure.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified time from Microsoft server"))
		return false
	}
	return metadataManager.Metadata.LastModified != lastModifiedDate.Unix()
}
