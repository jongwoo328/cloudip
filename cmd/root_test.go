package cmd

import (
	"bytes"
	"cloudip/common"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	savedVersion := Version
	t.Cleanup(func() { Version = savedVersion })
	Version = "0.8.1"

	cmd, _ := newTestCmd(t)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "0.8.1" {
		t.Errorf("expected '0.8.1', got '%s'", got)
	}
}

func TestVersionCmdRejectsArgs(t *testing.T) {
	cmd, _ := newTestCmd(t)
	cmd.SetArgs([]string{"version", "extra"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when passing args to version command")
	}
}

func TestFlagDefaults(t *testing.T) {
	cmd, _ := newTestCmd(t)
	flags := cmd.Flags()

	tests := []struct {
		name     string
		flag     string
		expected string
	}{
		{"format", "format", "text"},
		{"delimiter", "delimiter", " "},
		{"header", "header", "false"},
		{"verbose", "verbose", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag '%s' not found", tt.flag)
			}
			if f.DefValue != tt.expected {
				t.Errorf("expected default '%s', got '%s'", tt.expected, f.DefValue)
			}
		})
	}
}

func TestFlagBinding(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		verify func(t *testing.T, flags *common.CloudIpFlag)
	}{
		{
			name: "verbose short flag",
			args: []string{"-v"},
			verify: func(t *testing.T, flags *common.CloudIpFlag) {
				if !flags.Verbose {
					t.Error("expected Flags.Verbose to be true")
				}
			},
		},
		{
			name: "format flag",
			args: []string{"--format", "json"},
			verify: func(t *testing.T, flags *common.CloudIpFlag) {
				if flags.Format != "json" {
					t.Errorf("expected Flags.Format 'json', got '%s'", flags.Format)
				}
			},
		},
		{
			name: "delimiter flag",
			args: []string{"--delimiter", ","},
			verify: func(t *testing.T, flags *common.CloudIpFlag) {
				if flags.Delimiter != "," {
					t.Errorf("expected Flags.Delimiter ',', got '%s'", flags.Delimiter)
				}
			},
		},
		{
			name: "header flag",
			args: []string{"--header"},
			verify: func(t *testing.T, flags *common.CloudIpFlag) {
				if !flags.Header {
					t.Error("expected Flags.Header to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, flags := newTestCmd(t)
			err := cmd.Flags().Parse(tt.args)
			if err != nil {
				t.Fatalf("unexpected error parsing flags: %v", err)
			}
			tt.verify(t, flags)
		})
	}
}

func TestRootCmdReturnsErrorWhenAnyResultHasError(t *testing.T) {
	cmd, _ := newTestCmd(t)
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"bad-ip", "8.8.8.8"})

	var err error
	stdout := captureStdout(t, func() {
		err = cmd.Execute()
	})

	if err == nil {
		t.Fatal("expected command error for invalid IP")
	}
	if !strings.Contains(stdout, "bad-ip ERROR") {
		t.Fatalf("expected stdout to include error row, got %q", stdout)
	}
	if !strings.Contains(stdout, "8.8.8.8 unknown") {
		t.Fatalf("expected stdout to include successful unknown row, got %q", stdout)
	}
	if !strings.Contains(stderr.String(), "bad-ip: error parsing IP: bad-ip") {
		t.Fatalf("expected stderr to include detailed error, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr should not include usage for result errors, got %q", stderr.String())
	}
}

func TestRootCmdJSONErrorOutputIsValidJSON(t *testing.T) {
	cmd, _ := newTestCmd(t)
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"--format", "json", "bad-ip", "8.8.8.8"})

	var err error
	stdout := captureStdout(t, func() {
		err = cmd.Execute()
	})

	if err == nil {
		t.Fatal("expected command error for invalid IP")
	}

	var parsed []map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v, stdout: %q", err, stdout)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 JSON rows, got %d: %v", len(parsed), parsed)
	}
	if parsed[0]["ip"] != "bad-ip" || parsed[0]["provider"] != "error" || parsed[0]["error"] == "" {
		t.Fatalf("unexpected error row: %v", parsed[0])
	}
	if parsed[1]["ip"] != "8.8.8.8" || parsed[1]["provider"] != "unknown" || parsed[1]["error"] != "" {
		t.Fatalf("unexpected unknown row: %v", parsed[1])
	}
	if !strings.Contains(stderr.String(), "bad-ip: error parsing IP: bad-ip") {
		t.Fatalf("expected stderr to include detailed error, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr should not include usage for result errors, got %q", stderr.String())
	}
}
