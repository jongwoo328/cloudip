package cmd

import (
	"cloudip/common"
	"io"
	"os"
	"testing"

	"github.com/spf13/pflag"
)

// setupTest saves and restores all mutable global state shared across cmd tests:
// common.Flags, Version, rootCmd output writer, rootCmd args, and Cobra flag states.
// Call at the top of every test that touches any of these.
func setupTest(t *testing.T) {
	t.Helper()

	savedFlags := *common.Flags
	savedVersion := Version

	// Save Cobra flag values and Changed states
	type flagSnapshot struct {
		value   string
		changed bool
	}
	savedFlagStates := make(map[string]flagSnapshot)
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		savedFlagStates[f.Name] = flagSnapshot{
			value:   f.Value.String(),
			changed: f.Changed,
		}
	})

	t.Cleanup(func() {
		*common.Flags = savedFlags
		Version = savedVersion
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)

		rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
			if s, ok := savedFlagStates[f.Name]; ok {
				f.Value.Set(s.value)
				f.Changed = s.changed
			}
		})
	})

	// Reset to clean defaults
	*common.Flags = defaultTextFlags()
}

func defaultTextFlags() common.CloudIpFlag {
	return common.CloudIpFlag{
		Format:    "text",
		Delimiter: " ",
		Header:    false,
		Verbose:   false,
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	done := make(chan string)
	go func() {
		out, err := io.ReadAll(r)
		if err != nil {
			done <- ""
			return
		}
		done <- string(out)
	}()

	fn()
	w.Close()

	return <-done
}
