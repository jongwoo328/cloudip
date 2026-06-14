package aws

import (
	"cloudip/common"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAws,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.AWS,
		Signature: "",
	},
}

func (ipDataManagerAws *IpDataManagerAws) isExpired() (bool, error) {
	signature, err := ipDataManagerAws.GetSignatureUpstream()
	if err != nil {
		return false, err
	}
	return metadataManager.IsSignatureExpired(signature), nil
}
