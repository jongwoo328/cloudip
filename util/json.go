package util

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadJSON[T any](file *os.File, data *T) error {
	if err := json.NewDecoder(file).Decode(data); err != nil {
		return fmt.Errorf("failed to decode the data: %v", err)
	}
	return nil
}

func WriteJSON[T any](file *os.File, data *T) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode the data: %v", err)
	}
	return nil
}
