package cmd

import (
	"bytes"
	"cloudip/common"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// --- getProviderString ---

func TestGetProviderString(t *testing.T) {
	tests := []struct {
		name     string
		result   common.Result
		expected string
	}{
		{
			name:     "normal provider",
			result:   common.Result{Ip: "1.2.3.4", Provider: common.AWS},
			expected: "aws",
		},
		{
			name:     "empty provider returns unknown",
			result:   common.Result{Ip: "1.2.3.4", Provider: ""},
			expected: "unknown",
		},
		{
			name:     "error returns ERROR",
			result:   common.Result{Ip: "1.2.3.4", Error: fmt.Errorf("fail")},
			expected: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getProviderString(tt.result)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

// --- printResult dispatcher ---

func TestPrintResultDispatchesFormat(t *testing.T) {
	results := []common.Result{
		{Ip: "1.2.3.4", Provider: common.AWS},
	}

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "dispatches to text",
			format:   "text",
			expected: "1.2.3.4 aws",
		},
		{
			name:     "dispatches to json",
			format:   "json",
			expected: `[{"ip":"1.2.3.4","provider":"aws","error":""}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, flags := newTestCmd(t)
			flags.Format = tt.format

			var err error
			output := new(bytes.Buffer)
			err = printResult(output, results, flags)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := strings.TrimSpace(output.String())
			if got != tt.expected {
				t.Errorf("format '%s': expected %q, got %q", tt.format, tt.expected, got)
			}
		})
	}

	t.Run("dispatches to table", func(t *testing.T) {
		_, flags := newTestCmd(t)
		flags.Format = "table"

		var err error
		output := new(bytes.Buffer)
		err = printResult(output, results, flags)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(output.String(), "1.2.3.4") || !strings.Contains(output.String(), "aws") {
			t.Errorf("format 'table': expected output to contain '1.2.3.4' and 'aws', got %q", output.String())
		}
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		_, flags := newTestCmd(t)
		flags.Format = "yaml"

		output := new(bytes.Buffer)
		err := printResult(output, results, flags)

		if err == nil {
			t.Error("expected error for invalid format, got nil")
		}
		if !strings.Contains(err.Error(), "yaml") {
			t.Errorf("expected error to mention 'yaml', got: %v", err)
		}
	})
}

// --- Text format ---

func TestPrintResultAsText(t *testing.T) {
	tests := []struct {
		name      string
		delimiter string
		header    bool
		results   []common.Result
		expected  []string
	}{
		{
			name:      "basic output",
			delimiter: " ",
			results: []common.Result{
				{Ip: "1.2.3.4", Provider: common.AWS},
				{Ip: "5.6.7.8", Provider: ""},
			},
			expected: []string{"1.2.3.4 aws", "5.6.7.8 unknown"},
		},
		{
			name:      "with header",
			delimiter: " ",
			header:    true,
			results: []common.Result{
				{Ip: "1.2.3.4", Provider: common.AWS},
			},
			expected: []string{"IP Provider", "1.2.3.4 aws"},
		},
		{
			name:      "custom delimiter",
			delimiter: ",",
			results: []common.Result{
				{Ip: "1.2.3.4", Provider: common.GCP},
			},
			expected: []string{"1.2.3.4,gcp"},
		},
		{
			name:      "header uses delimiter",
			delimiter: ",",
			header:    true,
			results: []common.Result{
				{Ip: "1.2.3.4", Provider: common.AWS},
			},
			expected: []string{"IP,Provider", "1.2.3.4,aws"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, flags := newTestCmd(t)
			flags.Delimiter = tt.delimiter
			flags.Header = tt.header

			output := new(bytes.Buffer)
			if err := printResultAsText(output, tt.results, flags); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			if len(lines) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d: %q", len(tt.expected), len(lines), output.String())
			}
			for i, want := range tt.expected {
				if lines[i] != want {
					t.Errorf("line %d: expected '%s', got '%s'", i, want, lines[i])
				}
			}
		})
	}
}

// --- Table format ---

func TestPrintResultAsTable(t *testing.T) {
	_, flags := newTestCmd(t)

	results := []common.Result{
		{Ip: "1.2.3.4", Provider: common.AWS},
		{Ip: "5.6.7.8", Provider: common.Azure},
	}

	output := new(bytes.Buffer)
	if err := printResultAsTable(output, results, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")

	// header + 2 data rows
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 data), got %d: %q", len(lines), output.String())
	}

	// header is first line
	if !strings.Contains(lines[0], "IP") || !strings.Contains(lines[0], "Provider") {
		t.Errorf("expected header line to contain 'IP' and 'Provider', got: %q", lines[0])
	}

	// data rows contain correct values in order
	if !strings.Contains(lines[1], "1.2.3.4") || !strings.Contains(lines[1], "aws") {
		t.Errorf("expected first data row to contain '1.2.3.4' and 'aws', got: %q", lines[1])
	}
	if !strings.Contains(lines[2], "5.6.7.8") || !strings.Contains(lines[2], "azure") {
		t.Errorf("expected second data row to contain '5.6.7.8' and 'azure', got: %q", lines[2])
	}

	// header comes before data (IP should not appear in header)
	if strings.Contains(lines[0], "1.2.3.4") {
		t.Errorf("expected header not to contain data values, got: %q", lines[0])
	}
}

func TestPrintResultAsTableAlwaysHasHeader(t *testing.T) {
	_, flags := newTestCmd(t)
	flags.Header = false

	results := []common.Result{
		{Ip: "10.0.0.1", Provider: common.AWS},
	}

	output := new(bytes.Buffer)
	if err := printResultAsTable(output, results, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output.String(), "Provider") {
		t.Errorf("expected table to always include header, but 'Provider' not found in output: %q", output.String())
	}
}

// --- JSON format ---

func TestPrintResultAsJson(t *testing.T) {
	tests := []struct {
		name     string
		results  []common.Result
		expected []map[string]string
	}{
		{
			name: "basic with unknown",
			results: []common.Result{
				{Ip: "1.2.3.4", Provider: common.AWS},
				{Ip: "5.6.7.8", Provider: ""},
			},
			expected: []map[string]string{
				{"ip": "1.2.3.4", "provider": "aws", "error": ""},
				{"ip": "5.6.7.8", "provider": "unknown", "error": ""},
			},
		},
		{
			name: "multiple providers",
			results: []common.Result{
				{Ip: "1.1.1.1", Provider: common.AWS},
				{Ip: "2.2.2.2", Provider: common.GCP},
				{Ip: "3.3.3.3", Provider: common.Azure},
			},
			expected: []map[string]string{
				{"ip": "1.1.1.1", "provider": "aws", "error": ""},
				{"ip": "2.2.2.2", "provider": "gcp", "error": ""},
				{"ip": "3.3.3.3", "provider": "azure", "error": ""},
			},
		},
		{
			name: "error result",
			results: []common.Result{
				{Ip: "bad-ip", Error: fmt.Errorf("error parsing IP: bad-ip")},
			},
			expected: []map[string]string{
				{"ip": "bad-ip", "provider": "error", "error": "error parsing IP: bad-ip"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			if err := printResultAsJson(output, tt.results); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var parsed []map[string]string
			if err := json.Unmarshal([]byte(strings.TrimSpace(output.String())), &parsed); err != nil {
				t.Fatalf("output is not valid JSON: %v, output: %q", err, output.String())
			}

			if len(parsed) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d", len(tt.expected), len(parsed))
			}

			for i, want := range tt.expected {
				if parsed[i]["ip"] != want["ip"] || parsed[i]["provider"] != want["provider"] || parsed[i]["error"] != want["error"] {
					t.Errorf("item %d: expected %v, got %v", i, want, parsed[i])
				}
			}
		})
	}
}
