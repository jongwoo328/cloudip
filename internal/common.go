package internal

const AppName = "cloudip"

type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	GCP   CloudProvider = "gcp"
	Azure CloudProvider = "azure"
)

type CloudMetadata struct {
	Type         CloudProvider `json:"type"`
	LastModified int64         `json:"lastModified"`
}

type MetadataManager interface {
	EnsureDataFile() error // 데이터 파일이 없거나 오래된 경우 처리
	GetMetadata() (*CloudMetadata, error)
	WriteMetadata(metadata *CloudMetadata) error
	IsExpired() bool
}
