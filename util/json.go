package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadJSON[T any](file *os.File, data *T) error {
	return DecodeJSON(file, data)
}

func DecodeJSON[T any](reader io.Reader, data *T) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode the data: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return fmt.Errorf("invalid trailing JSON data: %w", err)
		}
		return fmt.Errorf("expected one JSON document")
	}
	return nil
}

func WriteJSON[T any](file *os.File, data *T) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode the data: %w", err)
	}
	return nil
}
