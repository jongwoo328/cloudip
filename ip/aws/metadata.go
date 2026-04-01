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

func init() {
	err := metadataManager.Ensure()
	if err != nil {
		return
	}

	err = metadataManager.Read()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error reading metadata"))
		return
	}
}

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAws.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified time from AWS server"))
		return false
	}
	return metadataManager.IsExpired(lastModifiedDate)
}
