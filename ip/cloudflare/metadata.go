package cloudflare

import (
	"cloudip/common"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathCloudflare,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.Cloudflare,
		Signature: "",
	},
}
