package cmd

import (
	"cloudip/common"
	"cloudip/ip"
	"cloudip/ip/provider"
	"encoding/json"
	"net"
	"os"
	"strings"
	"testing"
)

type verboseProvider struct{}

func (*verboseProvider) Initialize() error {
	common.VerboseOutput("Loading provider data")
	return nil
}

func (*verboseProvider) CheckParsedIP(net.IP) (bool, error) { return true, nil }
func (*verboseProvider) GetName() string                    { return "aws" }

func TestVerboseJSONSeparatesResultsAndDiagnostics(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		t.Run(map[bool]string{false: "quiet", true: "verbose"}[verbose], func(t *testing.T) {
			stderr, err := os.CreateTemp(t.TempDir(), "stderr")
			if err != nil {
				t.Fatal(err)
			}
			originalStderr := os.Stderr
			os.Stderr = stderr
			t.Cleanup(func() {
				os.Stderr = originalStderr
				stderr.Close()
				common.SetVerbose(false)
			})

			checker := ip.NewIPChecker(map[common.CloudProvider]provider.CloudProvider{
				common.AWS: &verboseProvider{},
			}, []common.CloudProvider{common.AWS})
			command := NewRootCmd(&common.CloudIpFlag{}, checker)
			args := []string{"1.2.3.4", "--format=json"}
			if verbose {
				args = append(args, "--verbose")
			}
			command.SetArgs(args)
			stdout := captureStdout(t, func() { err = command.Execute() })
			if err != nil {
				t.Fatal(err)
			}
			var results []jsonResult
			if err := json.Unmarshal([]byte(stdout), &results); err != nil {
				t.Fatalf("stdout must contain only JSON: %v; got %q", err, stdout)
			}
			if len(results) != 1 || results[0].Provider != "aws" {
				t.Fatalf("unexpected results: %+v", results)
			}
			logs, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			if verbose && !strings.Contains(string(logs), "Loading provider data") {
				t.Fatalf("missing diagnostics on stderr: %q", logs)
			}
			if !verbose && len(logs) != 0 {
				t.Fatalf("unexpected diagnostics without --verbose: %q", logs)
			}
		})
	}
}
