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
			fmt.Println("Error creating metadata file:", err)
			return fmt.Errorf("Error creating metadata file: %v", err)
		}
		defer func() {
			if err := metadataFile.Close(); err != nil {
				fmt.Println("Error closing metadata file:", err)
			}
		}()
		err = AwsMetadataManager.WriteMetadata(&internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: 0,
		})
		if err != nil {
			return fmt.Errorf("Error writing metadata: %v", err)
		}
	}

	if !util.IsFileExists(AwsMetadataManager.DataFilePath) {

	}
	return nil
}

func (AwsMetadataManager *MetadataManager) ReadMetadata() error {
	metadataFile, err := os.Open(AwsMetadataManager.MetadataFilePath)
	if err != nil {
		fmt.Println("Error opening metadata file:", err)
		return err
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			fmt.Println("Error closing metadata file:", err)
		}
	}()
	err = util.HandleJSON(metadataFile, AwsMetadataManager.Metadata, "read")
	if err != nil {
		fmt.Println("Error reading metadata file:", err)
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
		fmt.Println("Error opening metadata file:", err)
		return err
	}
	defer func() {
		if err := metadataFile.Close(); err != nil {
			fmt.Println("Error closing metadata file:", err)
		}
	}()
	if _, err := metadataFile.Seek(0, io.SeekStart); err != nil {
		fmt.Println("Error seeking metadata file:", err)
		return err
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
		fmt.Println("Error checking metadata file expiration:", err)
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Received non-200 status code:", resp.Status)
		return false
	}

	lastModified := resp.Header.Get("Last-Modified")
	lastModifiedDate, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		fmt.Println("Error parsing Date header:", err)
		return false
	}
	return lastModifiedDate.Unix() != AwsMetadataManager.Metadata.LastModified
}

func (AwsMetadataManager *MetadataManager) DownloadData() {
	resp, err := http.Get(AwsMetadataManager.DataURI)
	if err != nil {
		fmt.Println("Error downloading dataFile:", err)
		return
	}
	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Received non-200 status code:", resp.Status)
		return
	}

	dataFile, err := os.OpenFile(AwsMetadataManager.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening dataFile:", err)
		return
	}

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		fmt.Println("Error saving dataFile:", err)
		return
	}

	currentLastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		fmt.Println("Error parsing Date header:", err)
		return
	}
	err = AwsMetadataManager.ReadMetadata()
	if err != nil {
		fmt.Println("Error reading metadata !:", err)
		return
	}

	if AwsMetadataManager.Metadata.LastModified != currentLastModified.Unix() {
		metadata := internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: currentLastModified.Unix(),
		}
		if err := AwsMetadataManager.WriteMetadata(&metadata); err != nil {
			fmt.Println("Error writing metadata:", err)
			return
		}
	}

	defer func() {
		if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
			fmt.Println("Error closing response body:", networkCloseErr)
		}
		if fileCloseErr := dataFile.Close(); fileCloseErr != nil {
			fmt.Println("Error closing dataFile:", fileCloseErr)
		}
	}()

}
