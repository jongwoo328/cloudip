package azure

import (
	"cloudip/common"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAzure,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.Azure,
		Signature: "",
	},
}

func (ipDataManagerAzure *IpDataManagerAzure) isExpired() (bool, error) {
	lastModifiedDate, err := ipDataManagerAzure.GetLastModifiedUpstream()
	if err != nil {
		return false, err
	}
	return metadataManager.IsExpired(lastModifiedDate), nil
}
