package util

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"os"
)

func ReadJSON[T any](file *os.File, data *T) error {
	return DecodeJSON(file, data)
}

func DecodeJSON[T any](reader io.Reader, data *T) error {
	// UnmarshalRead consumes the entire reader and requires exactly one JSON value.
	if err := json.UnmarshalRead(reader, data); err != nil {
		return fmt.Errorf("failed to decode the data: %w", err)
	}
	return nil
}

func WriteJSON[T any](file *os.File, data *T) error {
	// Preserve cache formatting, including the encoder's terminating newline.
	encoder := jsontext.NewEncoder(file,
		jsontext.WithIndent("\t"),
		jsontext.EscapeForHTML(true),
		jsontext.EscapeForJS(true),
	)
	if err := json.MarshalEncode(encoder, data,
		json.Deterministic(true),
		json.FormatNilSliceAsNull(true),
		json.FormatNilMapAsNull(true),
	); err != nil {
		return fmt.Errorf("failed to encode the data: %w", err)
	}
	return nil
}
