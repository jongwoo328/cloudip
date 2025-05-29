package common_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ip-api/cloudip/common"
	// "github.com/ip-api/cloudip/util" // Not directly used here, but common package might.
)

func TestEnsureMetadataFile_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	providerDir := filepath.Join(tmpDir, "aws") // Using "aws" as a sample provider name
	metadataFilePath := filepath.Join(providerDir, ".metadata.json")
	initialMetadata := &common.CloudMetadata{
		Type:         common.AWS,
		LastModified: 1234567890,
	}

	err := common.EnsureMetadataFile(metadataFilePath, providerDir, common.AWS, initialMetadata)
	if err != nil {
		t.Fatalf("EnsureMetadataFile failed: %v", err)
	}

	// Verify provider directory was created
	if _, err := os.Stat(providerDir); os.IsNotExist(err) {
		t.Errorf("Provider directory %s was not created", providerDir)
	}

	// Verify metadata file was created
	if _, err := os.Stat(metadataFilePath); os.IsNotExist(err) {
		t.Errorf("Metadata file %s was not created", metadataFilePath)
	}

	// Verify initial metadata was written correctly
	readMeta := &common.CloudMetadata{}
	file, err := os.Open(metadataFilePath)
	if err != nil {
		t.Fatalf("Failed to open created metadata file: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(readMeta); err != nil {
		t.Fatalf("Failed to decode metadata from created file: %v", err)
	}

	if !reflect.DeepEqual(readMeta, initialMetadata) {
		t.Errorf("Written metadata %+v does not match initial metadata %+v", readMeta, initialMetadata)
	}
}

func TestEnsureMetadataFile_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	providerDir := filepath.Join(tmpDir, "gcp")
	metadataFilePath := filepath.Join(providerDir, ".metadata.json")

	// Pre-create the directory and a metadata file with different content
	if err := os.MkdirAll(providerDir, 0755); err != nil {
		t.Fatalf("Failed to create provider directory for test: %v", err)
	}
	existingMetadata := &common.CloudMetadata{
		Type:         common.GCP,
		LastModified: 9876543210,
	}
	file, err := os.Create(metadataFilePath)
	if err != nil {
		t.Fatalf("Failed to create existing metadata file: %v", err)
	}
	if err := json.NewEncoder(file).Encode(existingMetadata); err != nil {
		file.Close()
		t.Fatalf("Failed to encode existing metadata: %v", err)
	}
	file.Close()

	// Attempt to ensure the file again with different initial metadata
	differentInitialMetadata := &common.CloudMetadata{
		Type:         common.GCP,
		LastModified: 1111111111,
	}
	err = common.EnsureMetadataFile(metadataFilePath, providerDir, common.GCP, differentInitialMetadata)
	if err != nil {
		t.Fatalf("EnsureMetadataFile failed for existing file: %v", err)
	}

	// Verify the file content has NOT changed
	readMeta := &common.CloudMetadata{}
	readFile, err := os.Open(metadataFilePath)
	if err != nil {
		t.Fatalf("Failed to open metadata file after EnsureMetadataFile call: %v", err)
	}
	defer readFile.Close()

	if err := json.NewDecoder(readFile).Decode(readMeta); err != nil {
		t.Fatalf("Failed to decode metadata from file after EnsureMetadataFile call: %v", err)
	}

	if !reflect.DeepEqual(readMeta, existingMetadata) {
		t.Errorf("Existing metadata %+v was overwritten, expected %+v", readMeta, existingMetadata)
	}
}

func TestWriteAndReadMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	metadataFilePath := filepath.Join(tmpDir, "test_meta.json")

	originalMeta := &common.CloudMetadata{
		Type:         common.Azure,
		LastModified: 12345,
		ETag:         "test-etag",
		SyncToken:    "test-sync-token",
	}

	// Test WriteMetadata
	err := common.WriteMetadata(metadataFilePath, originalMeta)
	if err != nil {
		t.Fatalf("WriteMetadata failed: %v", err)
	}

	// Test ReadMetadata
	readMeta := &common.CloudMetadata{} // Target for reading
	err = common.ReadMetadata(metadataFilePath, readMeta)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	if !reflect.DeepEqual(readMeta, originalMeta) {
		t.Errorf("Read metadata %+v does not match original metadata %+v", readMeta, originalMeta)
	}
}

func TestReadMetadata_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	metadataFilePath := filepath.Join(tmpDir, "non_existent_meta.json")
	readMeta := &common.CloudMetadata{}

	err := common.ReadMetadata(metadataFilePath, readMeta)
	if err == nil {
		t.Error("ReadMetadata expected an error for non-existent file, got nil")
	}
}

func TestReadMetadata_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	metadataFilePath := filepath.Join(tmpDir, "empty_meta.json")

	// Create an empty file
	file, err := os.Create(metadataFilePath)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	file.Close()

	readMeta := &common.CloudMetadata{}
	err = common.ReadMetadata(metadataFilePath, readMeta)
	if err == nil {
		t.Error("ReadMetadata expected an error for empty file, got nil")
	}
	// More specific error check could be done here if desired (e.g., EOF or JSON syntax error)
}

func TestReadMetadata_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	metadataFilePath := filepath.Join(tmpDir, "malformed_meta.json")

	// Create a file with malformed JSON
	err := os.WriteFile(metadataFilePath, []byte("{this is not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write malformed JSON file: %v", err)
	}

	readMeta := &common.CloudMetadata{}
	err = common.ReadMetadata(metadataFilePath, readMeta)
	if err == nil {
		t.Error("ReadMetadata expected an error for malformed JSON, got nil")
	}
	// Check if it's a json.SyntaxError or similar
	if _, ok := err.(*json.SyntaxError); !ok {
		// common.ReadMetadata wraps the error, so we might not get SyntaxError directly
		// t.Errorf("Expected json.SyntaxError or wrapped version, got %T: %v", err, err)
		t.Logf("Note: Expected error for malformed JSON. Got type %T: %v", err, err)
	}
}

func TestWriteMetadata_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	// Path to a metadata file within a non-existent directory
	newProviderDir := filepath.Join(tmpDir, "new_provider")
	metadataFilePath := filepath.Join(newProviderDir, ".metadata.json")

	metadataToWrite := &common.CloudMetadata{
		Type:         common.AWS,
		LastModified: 7890,
	}

	// WriteMetadata should create 'new_provider' directory
	err := common.WriteMetadata(metadataFilePath, metadataToWrite)
	if err != nil {
		t.Fatalf("WriteMetadata failed when directory did not exist: %v", err)
	}

	// Verify file was created and content is correct
	if _, err := os.Stat(metadataFilePath); os.IsNotExist(err) {
		t.Errorf("Metadata file %s was not created by WriteMetadata", metadataFilePath)
	}

	readMeta := &common.CloudMetadata{}
	err = common.ReadMetadata(metadataFilePath, readMeta)
	if err != nil {
		t.Fatalf("Failed to read metadata written by WriteMetadata (dir creation test): %v", err)
	}
	if !reflect.DeepEqual(readMeta, metadataToWrite) {
		t.Errorf("Read metadata %+v does not match written metadata %+v (dir creation test)", readMeta, metadataToWrite)
	}
}
