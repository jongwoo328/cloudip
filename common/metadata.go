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
			return util.ErrorWithInfo(err, "error creating provider directory")
		}

		err := m.Write(&CloudMetadata{
			Type:         m.Metadata.Type,
			LastModified: 0,
		})
		if err != nil {
			return util.ErrorWithInfo(err, "error writing metadata")
		}
	}
	return nil
}

func (m *MetadataManager) Read() error {
	metadataFile, err := os.Open(m.MetadataFilePath)
	if err != nil {
		return util.ErrorWithInfo(err, "error opening metadata file")
	}
	defer metadataFile.Close()

	err = util.ReadJSON(metadataFile, m.Metadata)
	if err != nil {
		return util.ErrorWithInfo(err, "error reading metadata file")
	}
	return nil
}

func (m *MetadataManager) Write(metadata *CloudMetadata) error {
	metadataFile, err := os.OpenFile(m.MetadataFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return util.ErrorWithInfo(err, "error opening metadata file")
	}
	defer metadataFile.Close()

	err = util.WriteJSON(metadataFile, metadata)
	if err != nil {
		return util.ErrorWithInfo(err, "error writing metadata")
	}
	*m.Metadata = *metadata
	return nil
}

func (m *MetadataManager) IsExpired(upstreamLastModified time.Time) bool {
	return upstreamLastModified.Unix() != m.Metadata.LastModified
}
