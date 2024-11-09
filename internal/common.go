package internal

const AppName = "cloudip"

type CheckIpResult struct {
	Ip     string
	Result Result
}

type Result struct {
	Aws bool
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
