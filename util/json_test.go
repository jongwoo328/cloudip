package util

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestDecodeJSONStrictInput(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		valid       bool
	}{
		{"valid", `{"value":"ok"}`, true},
		{"whitespace", " \n{\"value\":\"ok\"}\n\t", true},
		{"unknown field", `{"value":"ok","future":true}`, true},
		{"duplicate known name", `{"value":"first","value":"last"}`, false},
		{"duplicate unknown name", `{"value":"ok","future":1,"future":2}`, false},
		{"nested duplicate", `{"future":{"key":1,"key":2}}`, false},
		{"invalid UTF-8", "{\"value\":\"\xff\"}", false},
		{"invalid UTF-8 in unknown field", "{\"future\":\"\xff\"}", false},
		{"empty input", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var data struct {
				Value string `json:"value"`
			}
			err := DecodeJSON(strings.NewReader(tc.input), &data)
			if (err == nil) != tc.valid {
				t.Fatalf("valid = %v, error = %v", tc.valid, err)
			}
			if tc.valid && data.Value != "ok" {
				t.Fatalf("decoded value = %q", data.Value)
			}
		})
	}
}

func TestDecodeJSONPreservesReaderErrors(t *testing.T) {
	failure := errors.New("response body interrupted")
	for _, prefix := range []string{`{"value":`, `{"value":"ok"}`, "{\"value\":\"ok\"}\n"} {
		reader := io.MultiReader(strings.NewReader(prefix), iotest.ErrReader(failure))
		var data map[string]string
		if err := DecodeJSON(reader, &data); !errors.Is(err, failure) {
			t.Fatalf("reader error lost for %q: %v", prefix, err)
		}
	}
}
