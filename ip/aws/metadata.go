package aws

import (
	"cloudip/common"
	"cloudip/util"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAws,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:         common.AWS,
		LastModified: 0,
	},
}

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAws.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting last modified time from AWS server"))
		return false
	}
	return metadataManager.IsExpired(lastModifiedDate)
}
