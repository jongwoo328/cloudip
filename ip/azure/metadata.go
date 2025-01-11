package azure

import (
	"cloudip/common"
	"cloudip/util"
	"fmt"
	"io"
	"os"
)

var metadataManager = &common.MetadataManager{
	MetadataFilePath: MetadataFilePathAzure,
	Metadata: &common.CloudMetadata{
		Type:         common.Azure,
		LastModified: 0,
	},
}

func init() {
	err := ensureMetadataFile()
	if err != nil {
		return
	}

	err = readMetadata()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error reading metadata"))
		return
	}
}

func ensureMetadataFile() error {
	if !util.IsFileExists(metadataManager.MetadataFilePath) {
		common.VerboseOutput(fmt.Sprintf("Creating %s ...", ProviderDirectory))
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			return util.ErrorWithInfo(err, "Error creating azure directory")
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
		err = writeMetadata(&common.CloudMetadata{
			Type:         common.Azure,
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

func writeMetadata(metadata *common.CloudMetadata) error {
	metadataFile, err := os.OpenFile(metadataManager.MetadataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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
	err = util.HandleJSON(metadataFile, metadata, "write")
	if err != nil {
		return util.ErrorWithInfo(err, "Error writing metadata")
	}
	err = readMetadata()
	return err
}

func isExpired() bool {
	lastModifiedDate, err := ipDataManagerAzure.GetLastModifiedUpstream()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error getting last modified time from Microsoft server"))
		return false
	}
	return metadataManager.Metadata.LastModified != lastModifiedDate.Unix()
}
