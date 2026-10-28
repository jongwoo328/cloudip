package oci

import (
	"cloudip/common"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathOCI,
	ProviderDir:      ProviderDirectory,
	Metadata: &common.CloudMetadata{
		Type:      common.OCI,
		Signature: "",
	},
}
