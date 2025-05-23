package aws

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
	// "io" // No longer used
	// "os" // No longer used
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAws,
	Metadata: &common.CloudMetadata{
		Type:         common.AWS,
		LastModified: 0,
	},
	// Add a SaveFunc that uses the generic WriteMetadata
	SaveFunc: func(filePath string, data *common.CloudMetadata) error {
		return common.WriteMetadata(filePath, data)
	},
}

func init() {
	initialMeta := &common.CloudMetadata{Type: common.AWS, LastModified: 0}
	err := common.EnsureMetadataFile(MetadataFilePathAws, ProviderDirectory, common.AWS, initialMeta)
	if err != nil {
		// EnsureMetadataFile already wraps errors, but PrintErrorTrace is good for top-level logging in init
		util.PrintErrorTrace(fmt.Errorf("Error ensuring AWS metadata file during init: %w", err))
		return
	}

	err = common.ReadMetadata(MetadataFilePathAws, metadataManager.Metadata)
	if err != nil {
		// ReadMetadata also wraps errors
		util.PrintErrorTrace(fmt.Errorf("Error reading AWS metadata during init: %w", err))
		return
	}
}

// ensureMetadataFile, readMetadata, and writeMetadata are removed,
// their functionality is replaced by common.EnsureMetadataFile,
// common.ReadMetadata, and common.WriteMetadata respectively.

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAws.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified time from AWS server"))
		return false
	}
	return lastModifiedDate.Unix() != metadataManager.Metadata.LastModified
}
