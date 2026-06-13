package azure

import (
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAzureDataURILoadsLazily(t *testing.T) {
	if ipDataManagerAzure.DataURI != "" {
		t.Fatal("Azure data URI should not be resolved during package initialization")
	}
}

func TestAzureDataURLClientHasTimeout(t *testing.T) {
	if dataURLClient.Timeout != 10*time.Second {
		t.Fatalf("expected data URL client timeout to be 10s, got %s", dataURLClient.Timeout)
	}
}

func TestAzureEnsureDataURIConcurrentAccess(t *testing.T) {
	manager := &IpDataManagerAzure{DataURI: "https://example.com/azure.json"}

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- manager.ensureDataURI()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("ensureDataURI returned unexpected error: %v", err)
		}
	}
}

func TestAzureLoadIpDataReturnsErrorForMissingFile(t *testing.T) {
	dir := t.TempDir()
	manager := &IpDataManagerAzure{
		DataFilePath: filepath.Join(dir, "missing.json"),
	}

	if _, err := manager.LoadIpData(); err == nil {
		t.Fatal("LoadIpData() error = nil, want error")
	}
}
