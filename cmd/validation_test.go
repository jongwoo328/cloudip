package cmd

import (
	"bytes"
	"cloudip/common"
	"cloudip/ip"
	"cloudip/ip/provider"
	"net"
	"strings"
	"testing"
)

type validationProvider struct{ calls int }

func (p *validationProvider) Initialize() error                { p.calls++; return nil }
func (*validationProvider) CheckParsedIP(net.IP) (bool, error) { return true, nil }
func (*validationProvider) GetName() string                    { return "aws" }

func TestArgumentValidationBeforeProviderLookup(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		wantError string
	}{
		{"missing IP", []string{}, "Usage: cloudip IP [IP...]"},
		{"flags without IP", []string{"--format=json", "--verbose"}, "Usage:"},
		{"invalid format", []string{"1.2.3.4", "--format=typo"}, "invalid output format: typo"},
		{"empty format", []string{"1.2.3.4", "--format="}, "invalid output format:"},
		{"help", []string{"--help"}, ""},
		{"version", []string{"version"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &validationProvider{}
			checker := ip.NewIPChecker(map[common.CloudProvider]provider.CloudProvider{common.AWS: p}, []common.CloudProvider{common.AWS})
			command := NewRootCmd(&common.CloudIpFlag{}, checker)
			var stdout, stderr bytes.Buffer
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			command.SetArgs(tc.args)
			err := command.Execute()
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("error = %v, want %q", err, tc.wantError)
				}
				if stdout.Len() != 0 {
					t.Fatalf("validation wrote result output: %q", stdout.String())
				}
			}
			if p.calls != 0 {
				t.Fatalf("provider initialized %d times for %v", p.calls, tc.args)
			}
		})
	}
}

func TestSupportedFormatsStillPerformLookup(t *testing.T) {
	for _, format := range []string{"text", "table", "json"} {
		t.Run(format, func(t *testing.T) {
			p := &validationProvider{}
			checker := ip.NewIPChecker(map[common.CloudProvider]provider.CloudProvider{common.AWS: p}, []common.CloudProvider{common.AWS})
			command := NewRootCmd(&common.CloudIpFlag{}, checker)
			var stdout bytes.Buffer
			command.SetOut(&stdout)
			command.SetArgs([]string{"1.2.3.4", "--format=" + format})
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if p.calls != 1 || !strings.Contains(stdout.String(), "aws") {
				t.Fatalf("calls=%d, stdout=%q", p.calls, stdout.String())
			}
		})
	}
}
