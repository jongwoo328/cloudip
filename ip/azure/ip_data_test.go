package azure

import (
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
