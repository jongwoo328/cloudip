package azure

import (
	"cloudip/common"
	"cloudip/util"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAzure,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:         common.Azure,
		LastModified: 0,
	},
}

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAzure.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting last modified time from Microsoft server"))
		return false
	}
	return metadataManager.IsExpired(lastModifiedDate)
}
