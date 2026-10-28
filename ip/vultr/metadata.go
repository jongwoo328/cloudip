package vultr

import (
	"cloudip/common"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathVultr,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.Vultr,
		Signature: "",
	},
}
