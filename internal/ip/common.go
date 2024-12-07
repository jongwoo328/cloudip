package ip

import "cloudip/internal"

var Providers = []internal.CloudProvider{
	"aws",
}

type MetadataManager interface {
	EnsureMetadataFile() error // 데이터 파일이 없거나 오래된 경우 처리
	GetMetadata() (*internal.CloudMetadata, error)
	WriteMetadata(metadata *internal.CloudMetadata) error
	IsExpired() bool
}
