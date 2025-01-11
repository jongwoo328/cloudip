package common

type CheckIpResult struct {
	Ip     string
	Result Result
	Error  error
}

type Result struct {
	Aws   bool
	Gcp   bool
	Azure bool
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
	Metadata         *CloudMetadata
}
