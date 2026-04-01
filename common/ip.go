package common

type CheckIpResult struct {
	Ip       string
	Provider CloudProvider
	Error    error
}

const (
	AWS   CloudProvider = "aws"
	GCP   CloudProvider = "gcp"
	Azure CloudProvider = "azure"
)

type CloudProvider string

type CloudMetadata struct {
	Type         CloudProvider `json:"type"`
	LastModified int64         `json:"lastModified"`
}

type MetadataManager struct {
	MetadataFilePath string
	ProviderDir      string
	Metadata         *CloudMetadata
}
