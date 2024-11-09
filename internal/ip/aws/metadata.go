package aws

import (
	"cloudip/internal"
	"cloudip/internal/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type MetadataManager struct {
	DataFilePath     string
	MetadataFilePath string
	Metadata         *internal.CloudMetadata
	DataURI          string
}

var metadataManager = &MetadataManager{
	DataFilePath:     DataFilePath,
	MetadataFilePath: MetadataFilePath,
	Metadata: &internal.CloudMetadata{
		Type:         internal.AWS,
		LastModified: 0,
	},
	DataURI: DataUrl,
}

func GetMetadataManager() *MetadataManager {
	return metadataManager
}

func (AwsMetadataManager *MetadataManager) EnsureDataFile() error {
	if !util.IsFileExists(AwsMetadataManager.MetadataFilePath) {
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

	if !util.IsFileExists(AwsMetadataManager.DataFilePath) {

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
	err := AwsMetadataManager.ReadMetadata()
	if err != nil {
		return true
	}

	resp, err := http.Head(AwsMetadataManager.DataURI)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error checking metadata file expiration"))
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing response body"))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		util.PrintErrorTrace(util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error checking metadata file expiration"))
		return false
	}

	lastModified := resp.Header.Get("Last-Modified")
	lastModifiedDate, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error parsing Date header"))
		return false
	}
	return lastModifiedDate.Unix() != AwsMetadataManager.Metadata.LastModified
}

func (AwsMetadataManager *MetadataManager) DownloadData() {
	resp, err := http.Get(AwsMetadataManager.DataURI)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error downloading dataFile"))
		return
	}
	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		util.PrintErrorTrace(util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error downloading dataFile"))
		return
	}

	dataFile, err := os.OpenFile(AwsMetadataManager.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error opening dataFile"))
		return
	}

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error saving dataFile"))
		return
	}

	currentLastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error parsing Date header"))
		return
	}
	err = AwsMetadataManager.ReadMetadata()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error reading metadata"))
		return
	}

	if AwsMetadataManager.Metadata.LastModified != currentLastModified.Unix() {
		metadata := internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: currentLastModified.Unix(),
		}
		if err := AwsMetadataManager.WriteMetadata(&metadata); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error writing metadata"))
			return
		}
	}

	defer func() {
		if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(networkCloseErr, "Error closing response body"))
		}
		if fileCloseErr := dataFile.Close(); fileCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(fileCloseErr, "Error closing dataFile"))
		}
	}()

}
