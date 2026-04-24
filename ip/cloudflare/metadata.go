package cloudflare

import (
	"cloudip/common"
	"cloudip/util"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathCloudflare,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.Cloudflare,
		Signature: "",
	},
}

func isExpired() bool {
	signature, err := ipDataManagerCloudflare.GetSignatureUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting signature from Cloudflare"))
		return false
	}
	return metadataManager.IsSignatureExpired(signature)
}
