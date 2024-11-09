package factory

import (
	"cloudip/internal"
	"cloudip/internal/ip/aws"
	"fmt"
)

func NewMetadataManager(provider internal.CloudProvider) (internal.MetadataManager, error) {
	switch provider {
	case internal.AWS:
		return aws.NewAwsMetadataManager(), nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}
