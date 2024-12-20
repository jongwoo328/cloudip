package gcp

import (
	"cloudip/internal"
	"cloudip/internal/util"
	"io"
	"os"
)

var metadataManager = &internal.MetadataManager{
	MetadataFilePath: MetadataFilePathAws,
	Metadata: &internal.CloudMetadata{
		Type:         internal.GCP,
		LastModified: 0,
	},
}

func ensureMetadataFile() error {
	if !util.IsFileExists(metadataManager.MetadataFilePath) {
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			return util.ErrorWithInfo(err, "Error creating gcp directory")
		}
		metadataFile, err := os.Create(metadataManager.MetadataFilePath)
		if err != nil {
			return util.ErrorWithInfo(err, "Error creating metadata file")
		}
		defer func() {
			if err := metadataFile.Close(); err != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing metadata file"))
			}
		}()
		err = writeMetadata(&internal.CloudMetadata{
			Type:         internal.GCP,
			LastModified: 0,
		})
		if err != nil {
			return util.ErrorWithInfo(err, "Error writing metadata")
		}
	}

	return nil
}

func readMetadata() error {
	metadataFile, err := os.Open(metadataManager.MetadataFilePath)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error opening metadata file"))
		return err
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing metadata file"))
		}
	}()
	err = util.HandleJSON(metadataFile, metadataManager.Metadata, "read")
	if err != nil {
		err = util.ErrorWithInfo(err, "Error reading metadata file")
		return err
	}
	return nil
}

func writeMetadata(metadata *internal.CloudMetadata) error {
	metadataFile, err := os.Create(metadataManager.MetadataFilePath)
	if err != nil {
		return util.ErrorWithInfo(err, "Error creating metadata file")
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

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerGcp.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified date"))
		return false
	}
	return lastModifiedDate.Unix() != metadataManager.Metadata.LastModified
}
