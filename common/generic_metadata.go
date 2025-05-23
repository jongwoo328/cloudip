package common

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ip-api/cloudip/util"
)

// EnsureMetadataFile checks if the metadata file exists. If not, it creates the directory
// and the initial metadata file.
func EnsureMetadataFile(metadataFilePath string, providerDir string, providerType CloudProviderType, initialMetadata *CloudMetadata) error {
	if _, err := os.Stat(metadataFilePath); os.IsNotExist(err) {
		VerboseOutput(fmt.Sprintf("Creating directory: %s", providerDir))
		if mkdirErr := os.MkdirAll(providerDir, 0755); mkdirErr != nil {
			return util.WrapError(mkdirErr, fmt.Sprintf("failed to create directory: %s", providerDir))
		}

		VerboseOutput(fmt.Sprintf("Creating initial metadata file for %s: %s", providerType, metadataFilePath))
		if writeErr := WriteMetadata(metadataFilePath, initialMetadata); writeErr != nil {
			return util.WrapError(writeErr, fmt.Sprintf("failed to write initial metadata file for %s", providerType))
		}
		return nil
	} else if err != nil {
		return util.WrapError(err, fmt.Sprintf("failed to stat metadata file: %s", metadataFilePath))
	}
	return nil
}

// ReadMetadata reads and unmarshals the JSON content from the metadata file.
func ReadMetadata(metadataFilePath string, metadata *CloudMetadata) error {
	file, err := os.Open(metadataFilePath)
	if err != nil {
		return util.WrapError(err, fmt.Sprintf("failed to open metadata file for reading: %s", metadataFilePath))
	}
	defer file.Close()

	err = util.HandleJSON(nil, json.NewDecoder(file).Decode(&metadata), "decode metadata from file: "+metadataFilePath)
	if err != nil {
		return err // HandleJSON already wraps the error
	}
	return nil
}

// WriteMetadata marshals and writes the metadata struct to the specified file as JSON.
func WriteMetadata(metadataFilePath string, metadata *CloudMetadata) error {
	// Ensure the directory exists before trying to write the file
	dir := filepath.Dir(metadataFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
			return util.WrapError(mkdirErr, fmt.Sprintf("failed to create directory for metadata file: %s", dir))
		}
	}

	file, err := os.OpenFile(metadataFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return util.WrapError(err, fmt.Sprintf("failed to open metadata file for writing: %s", metadataFilePath))
	}
	defer file.Close()

	// Seek to the beginning of the file before writing
	_, seekErr := file.Seek(0, io.SeekStart)
	if seekErr != nil {
		return util.WrapError(seekErr, fmt.Sprintf("failed to seek to beginning of metadata file: %s", metadataFilePath))
	}

	// Truncate the file to remove old content if any, especially after seeking.
	// This is important if the new content is smaller than the old.
	if truncErr := file.Truncate(0); truncErr != nil {
		return util.WrapError(truncErr, fmt.Sprintf("failed to truncate metadata file: %s", metadataFilePath))
	}


	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // For pretty printing, consistent with existing logic
	err = util.HandleJSON(nil, encoder.Encode(metadata), "encode metadata to file: "+metadataFilePath)
	if err != nil {
		return err // HandleJSON already wraps the error
	}
	return nil
}
