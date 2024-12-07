package aws

import (
	"cloudip/internal"
	"cloudip/internal/util"
	"io"
	"os"
)

type MetadataManager struct {
	MetadataFilePath string
	Metadata         *internal.CloudMetadata
}

var metadataManager = &MetadataManager{
	MetadataFilePath: MetadataFilePath,
	Metadata: &internal.CloudMetadata{
		Type:         internal.AWS,
		LastModified: 0,
	},
}

func GetMetadataManager() *MetadataManager {
	return metadataManager
}

func (AwsMetadataManager *MetadataManager) EnsureMetadataFile() error {
	if !util.IsFileExists(AwsMetadataManager.MetadataFilePath) {
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			return util.ErrorWithInfo(err, "Error creating aws directory")
		}
		metadataFile, err := os.Create(AwsMetadataManager.MetadataFilePath)
		if err != nil {
			return util.ErrorWithInfo(err, "Error creating metadata file")
		}
		defer func() {
			if err := metadataFile.Close(); err != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing metadata file"))
			}
		}()
		err = AwsMetadataManager.WriteMetadata(&internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: 0,
		})
		if err != nil {
			return util.ErrorWithInfo(err, "Error writing metadata")
		}
	}

	return nil
}

func (AwsMetadataManager *MetadataManager) ReadMetadata() error {
	metadataFile, err := os.Open(AwsMetadataManager.MetadataFilePath)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error opening metadata file"))
		return err
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing metadata file"))
		}
	}()
	err = util.HandleJSON(metadataFile, AwsMetadataManager.Metadata, "read")
	if err != nil {
		err = util.ErrorWithInfo(err, "Error reading metadata file")
		return err
	}
	return nil
}

func (AwsMetadataManager *MetadataManager) GetMetadata() (*internal.CloudMetadata, error) {
	err := AwsMetadataManager.ReadMetadata()
	if err != nil {
		return nil, err
	}
	return AwsMetadataManager.Metadata, nil
}

func (AwsMetadataManager *MetadataManager) WriteMetadata(metadata *internal.CloudMetadata) error {
	metadataFile, err := os.OpenFile(AwsMetadataManager.MetadataFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return util.ErrorWithInfo(err, "Error opening metadata file")
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing metadata file"))
		}
	}()
	if _, err := metadataFile.Seek(0, io.SeekStart); err != nil {
		return util.ErrorWithInfo(err, "Error seeking metadata file")
	}
	return util.HandleJSON(metadataFile, metadata, "write")
}

func (AwsMetadataManager *MetadataManager) IsExpired() bool {
	ipDataManagerAws := GetIpDataManagerAws()
	lastModifiedDate, err := ipDataManagerAws.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified date"))
		return false
	}
	return lastModifiedDate.Unix() != AwsMetadataManager.Metadata.LastModified
}
