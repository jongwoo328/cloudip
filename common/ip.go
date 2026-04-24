package common

type Result struct {
	Ip       string
	Provider CloudProvider
	Error    error
}

const (
	AWS        CloudProvider = "aws"
	GCP        CloudProvider = "gcp"
	Azure      CloudProvider = "azure"
	Cloudflare CloudProvider = "cloudflare"
)

type CloudProvider string

type CloudMetadata struct {
	Type      CloudProvider `json:"type"`
	Signature string        `json:"signature"`
}

type MetadataManager struct {
	MetadataFilePath string
	ProviderDir      string
	Metadata         *CloudMetadata
}
