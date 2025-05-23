package gcp

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
	// "io" // No longer used
	// "os" // No longer used
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathGcp,
	Metadata: &common.CloudMetadata{
		Type:         common.GCP,
		LastModified: 0,
	},
	// Add a SaveFunc that uses the generic WriteMetadata
	SaveFunc: func(filePath string, data *common.CloudMetadata) error {
		return common.WriteMetadata(filePath, data)
	},
}

func init() {
	initialMeta := &common.CloudMetadata{Type: common.GCP, LastModified: 0}
	err := common.EnsureMetadataFile(MetadataFilePathGcp, ProviderDirectory, common.GCP, initialMeta)
	if err != nil {
		util.PrintErrorTrace(fmt.Errorf("Error ensuring GCP metadata file during init: %w", err))
		return
	}

	err = common.ReadMetadata(MetadataFilePathGcp, metadataManager.Metadata)
	if err != nil {
		util.PrintErrorTrace(fmt.Errorf("Error reading GCP metadata during init: %w", err))
		return
	}
}

// ensureMetadataFile, readMetadata, and writeMetadata are removed,
// their functionality is replaced by common.EnsureMetadataFile,
// common.ReadMetadata, and common.WriteMetadata respectively.

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerGcp.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified time from GCP server"))
		return false
	}
	return lastModifiedDate.Unix() != metadataManager.Metadata.LastModified
}
