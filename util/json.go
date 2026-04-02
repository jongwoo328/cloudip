package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadJSON[T any](file *os.File, data *T) error {
	readData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read the file: %v", err)
	}
	if err := json.Unmarshal(readData, data); err != nil {
		return fmt.Errorf("failed to unmarshal the data: %v", err)
	}
	return nil
}

func WriteJSON[T any](file *os.File, data *T) error {
	writeData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal the data: %v", err)
	}
	if _, err := file.Write(writeData); err != nil {
		return fmt.Errorf("failed to write the data: %v", err)
	}
	return nil
}
