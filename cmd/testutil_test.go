package cmd

import (
	"cloudip/common"
	"cloudip/ip"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func newTestCmd(t *testing.T) (*cobra.Command, *common.CloudIpFlag) {
	t.Helper()
	flags := &common.CloudIpFlag{}
	checker := ip.NewIPChecker(nil, nil)
	cmd := NewRootCmd(flags, checker)
	return cmd, flags
}

type captureResult struct {
	output string
	err    error
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

	done := make(chan captureResult)
	go func() {
		out, err := io.ReadAll(r)
		done <- captureResult{output: string(out), err: err}
	}()

	fn()
	w.Close()

	result := <-done
	if result.err != nil {
		t.Fatalf("failed to read captured stdout: %v", result.err)
	}
	return result.output
}
