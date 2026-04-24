package gcp

import (
	"cloudip/common"
	"cloudip/util"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathGcp,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.GCP,
		Signature: "",
	},
}

func isExpired() bool {
	signature, err := ipDataManagerGcp.GetSignatureUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting signature from GCP server"))
		return false
	}
	return metadataManager.IsSignatureExpired(signature)
}
