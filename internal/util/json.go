package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
