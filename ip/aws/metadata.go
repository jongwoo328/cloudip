package aws

import (
	"cloudip/common"
	util2 "cloudip/util"
	"io"
	"os"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAws,
	Metadata: &common.CloudMetadata{
		Type:         common.AWS,
		LastModified: 0,
	},
}

func ensureMetadataFile() error {
	if !util2.IsFileExists(metadataManager.MetadataFilePath) {
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			return util2.ErrorWithInfo(err, "Error creating aws directory")
		}
		metadataFile, err := os.Create(metadataManager.MetadataFilePath)
		if err != nil {
			return util2.ErrorWithInfo(err, "Error creating metadata file")
		}
		defer func() {
			if err := metadataFile.Close(); err != nil {
				util2.PrintErrorTrace(util2.ErrorWithInfo(err, "Error closing metadata file"))
			}
		}()
		err = writeMetadata(&common.CloudMetadata{
			Type:         common.AWS,
			LastModified: 0,
		})
		if err != nil {
			return util2.ErrorWithInfo(err, "Error writing metadata")
		}
	}

	return nil
}

func readMetadata() error {
	metadataFile, err := os.Open(metadataManager.MetadataFilePath)
	if err != nil {
		util2.PrintErrorTrace(util2.ErrorWithInfo(err, "Error opening metadata file"))
		return err
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			util2.PrintErrorTrace(util2.ErrorWithInfo(err, "Error closing metadata file"))
		}
	}()
	err = util2.HandleJSON(metadataFile, metadataManager.Metadata, "read")
	if err != nil {
		err = util2.ErrorWithInfo(err, "Error reading metadata file")
		return err
	}
	return nil
}

func writeMetadata(metadata *common.CloudMetadata) error {
	metadataFile, err := os.OpenFile(metadataManager.MetadataFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return util2.ErrorWithInfo(err, "Error opening metadata file")
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			util2.PrintErrorTrace(util2.ErrorWithInfo(err, "Error closing metadata file"))
		}
	}()
	if _, err := metadataFile.Seek(0, io.SeekStart); err != nil {
		return util2.ErrorWithInfo(err, "Error seeking metadata file")
	}
	return util2.HandleJSON(metadataFile, metadata, "write")
}

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAws.GetLastModifiedUpstream()
	if err != nil {
		util2.PrintErrorTrace(util2.ErrorWithInfo(err, "Error getting last modified date"))
		return false
	}
	return lastModifiedDate.Unix() != metadataManager.Metadata.LastModified
}
