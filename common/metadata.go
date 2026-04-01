package common

import (
	"cloudip/util"
	"fmt"
	"os"
	"time"
)

func (m *MetadataManager) Ensure() error {
	if !util.IsFileExists(m.MetadataFilePath) {
		VerboseOutput(fmt.Sprintf("Creating %s ...", m.ProviderDir))
		if err := os.MkdirAll(m.ProviderDir, 0755); err != nil {
			return util.ErrorWithInfo(err, "Error creating provider directory")
		}
		metadataFile, err := os.Create(m.MetadataFilePath)
		if err != nil {
			return util.ErrorWithInfo(err, "Error creating metadata file")
		}
		defer metadataFile.Close()

		err = m.Write(&CloudMetadata{
			Type:         m.Metadata.Type,
			LastModified: 0,
		})
		if err != nil {
			return util.ErrorWithInfo(err, "Error writing metadata")
		}
	}
	return nil
}

func (m *MetadataManager) Read() error {
	metadataFile, err := os.Open(m.MetadataFilePath)
	if err != nil {
		return util.ErrorWithInfo(err, "Error opening metadata file")
	}
	defer metadataFile.Close()

	err = util.HandleJSON(metadataFile, m.Metadata, "read")
	if err != nil {
		return util.ErrorWithInfo(err, "Error reading metadata file")
	}
	return nil
}

func (m *MetadataManager) Write(metadata *CloudMetadata) error {
	metadataFile, err := os.OpenFile(m.MetadataFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return util.ErrorWithInfo(err, "Error opening metadata file")
	}
	defer metadataFile.Close()

	err = util.HandleJSON(metadataFile, metadata, "write")
	if err != nil {
		return util.ErrorWithInfo(err, "Error writing metadata")
	}
	return m.Read()
}

func (m *MetadataManager) IsExpired(upstreamLastModified time.Time) bool {
	return upstreamLastModified.Unix() != m.Metadata.LastModified
}
