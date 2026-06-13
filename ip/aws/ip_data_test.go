package aws

import (
	"path/filepath"
	"testing"
)

func TestAWSLoadIpDataReturnsErrorForMissingFile(t *testing.T) {
	dir := t.TempDir()
	manager := &IpDataManagerAws{
		DataFilePath: filepath.Join(dir, "missing.json"),
	}

	if _, err := manager.LoadIpData(); err == nil {
		t.Fatal("LoadIpData() error = nil, want error")
	}
}
