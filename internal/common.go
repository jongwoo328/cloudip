package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const AppName = "cloudip"

type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	GCP   CloudProvider = "gcp"
	Azure CloudProvider = "azure"
)

type CloudMetadata struct {
	Type         CloudProvider `json:"type"`
	LastModified int64         `json:"lastModified"`
}

type MetadataManager interface {
	EnsureDataFile() error // 데이터 파일이 없거나 오래된 경우 처리
	GetMetadata() (*CloudMetadata, error)
	WriteMetadata(metadata *CloudMetadata) error
	IsExpired() bool
}

func HandleJSON[T any](file *os.File, data *T, mode string) error {
	switch mode {
	case "read":
		readData, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read the file: %v", err)
		}
		if err := json.Unmarshal(readData, data); err != nil {
			return fmt.Errorf("failed to unmarshal the data: %v", err)
		}
	case "write":
		writeData, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return fmt.Errorf("failed to marshal the data: %v", err)
		}
		if _, err := file.Write(writeData); err != nil {
			return fmt.Errorf("failed to write the data: %v", err)
		}
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	return nil
}
