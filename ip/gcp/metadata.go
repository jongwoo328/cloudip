package gcp

import (
	"cloudip/common"
	"cloudip/util"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathGcp,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:         common.GCP,
		LastModified: 0,
	},
}

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerGcp.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting last modified time from GCP server"))
		return false
	}
	return metadataManager.IsExpired(lastModifiedDate)
}
