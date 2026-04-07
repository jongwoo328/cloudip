package cmd

import (
	"bytes"
	"cloudip/common"
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
