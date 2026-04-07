package cmd

import (
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
		name      string
		format    string
		delimiter string
		expected  string
	}{
		{
			name:      "dispatches to text",
			format:    "text",
			delimiter: " ",
			expected:  "1.2.3.4 aws",
		},
		{
			name:      "dispatches to json",
			format:    "json",
			delimiter: " ",
			expected:  `[{"IP":"1.2.3.4","Provider":"aws"}]`,
		},
		{
			name:      "dispatches to table",
			format:    "table",
			delimiter: "\t",
			expected:  "1.2.3.4\taws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			common.Flags.Format = tt.format
			common.Flags.Delimiter = tt.delimiter

			output := captureStdout(t, func() {
				printResult(&results)
			})

			got := strings.TrimSpace(output)
			if !strings.Contains(got, tt.expected) {
				t.Errorf("format '%s': expected output to contain %q, got %q", tt.format, tt.expected, got)
			}
		})
	}
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
			setupTest(t)
			common.Flags.Delimiter = tt.delimiter
			common.Flags.Header = tt.header

			output := captureStdout(t, func() {
				printResultAsText(&tt.results)
			})

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d: %q", len(tt.expected), len(lines), output)
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
	setupTest(t)

	results := []common.Result{
		{Ip: "1.2.3.4", Provider: common.AWS},
		{Ip: "5.6.7.8", Provider: common.Azure},
	}

	output := captureStdout(t, func() {
		printResultAsTable(&results)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header + 2 data), got %d: %q", len(lines), output)
	}
	if !strings.Contains(lines[0], "IP") || !strings.Contains(lines[0], "Provider") {
		t.Errorf("expected header line to contain 'IP' and 'Provider', got: %q", lines[0])
	}
	if !strings.Contains(output, "1.2.3.4") || !strings.Contains(output, "aws") {
		t.Errorf("expected table to contain '1.2.3.4' and 'aws', got: %q", output)
	}
	if !strings.Contains(output, "5.6.7.8") || !strings.Contains(output, "azure") {
		t.Errorf("expected table to contain '5.6.7.8' and 'azure', got: %q", output)
	}
}

func TestPrintResultAsTableAlwaysHasHeader(t *testing.T) {
	setupTest(t)
	common.Flags.Header = false

	results := []common.Result{
		{Ip: "10.0.0.1", Provider: common.AWS},
	}

	output := captureStdout(t, func() {
		printResultAsTable(&results)
	})

	if !strings.Contains(output, "Provider") {
		t.Errorf("expected table to always include header, but 'Provider' not found in output: %q", output)
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
				{"IP": "1.2.3.4", "Provider": "aws"},
				{"IP": "5.6.7.8", "Provider": "unknown"},
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
				{"IP": "1.1.1.1", "Provider": "aws"},
				{"IP": "2.2.2.2", "Provider": "gcp"},
				{"IP": "3.3.3.3", "Provider": "azure"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)

			output := captureStdout(t, func() {
				printResultAsJson(&tt.results)
			})

			var parsed []map[string]string
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
				t.Fatalf("output is not valid JSON: %v, output: %q", err, output)
			}

			if len(parsed) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d", len(tt.expected), len(parsed))
			}

			for i, want := range tt.expected {
				if parsed[i]["IP"] != want["IP"] || parsed[i]["Provider"] != want["Provider"] {
					t.Errorf("item %d: expected %v, got %v", i, want, parsed[i])
				}
			}
		})
	}
}
