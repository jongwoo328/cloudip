package cmd

import (
	"bytes"
	"cloudip/common"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	setupTest(t)
	Version = "0.8.1"

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	if got != "0.8.1" {
		t.Errorf("expected '0.8.1', got '%s'", got)
	}
}

func TestVersionCmdRejectsArgs(t *testing.T) {
	setupTest(t)
	rootCmd.SetArgs([]string{"version", "extra"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when passing args to version command")
	}
}

func TestFlagDefaults(t *testing.T) {
	flags := rootCmd.Flags()

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
		verify func(t *testing.T)
	}{
		{
			name: "verbose short flag",
			args: []string{"-v"},
			verify: func(t *testing.T) {
				if !common.Flags.Verbose {
					t.Error("expected Flags.Verbose to be true")
				}
			},
		},
		{
			name: "format flag",
			args: []string{"--format", "json"},
			verify: func(t *testing.T) {
				if common.Flags.Format != "json" {
					t.Errorf("expected Flags.Format 'json', got '%s'", common.Flags.Format)
				}
			},
		},
		{
			name: "delimiter flag",
			args: []string{"--delimiter", ","},
			verify: func(t *testing.T) {
				if common.Flags.Delimiter != "," {
					t.Errorf("expected Flags.Delimiter ',', got '%s'", common.Flags.Delimiter)
				}
			},
		},
		{
			name: "header flag",
			args: []string{"--header"},
			verify: func(t *testing.T) {
				if !common.Flags.Header {
					t.Error("expected Flags.Header to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			err := rootCmd.Flags().Parse(tt.args)
			if err != nil {
				t.Fatalf("unexpected error parsing flags: %v", err)
			}
			tt.verify(t)
		})
	}
}
