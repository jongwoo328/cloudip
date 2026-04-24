package aws

import (
	"cloudip/common"
	"cloudip/util"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAws,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.AWS,
		Signature: "",
	},
}

func isExpired() bool {
	signature, err := ipDataManagerAws.GetSignatureUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting signature from AWS server"))
		return false
	}
	return metadataManager.IsSignatureExpired(signature)
}
