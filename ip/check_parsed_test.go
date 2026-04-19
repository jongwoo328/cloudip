package ip

import (
	"cloudip/common"
	"cloudip/ip/provider"
	"net"
	"testing"
)

type parsedPathMockProvider struct {
	name             string
	initializeCalls  int
	checkParsedCalls int
	initErr          error
	checkParsedErr   error
	parsedMatch      bool
}

func (m *parsedPathMockProvider) Initialize() error {
	m.initializeCalls++
	return m.initErr
}

func (m *parsedPathMockProvider) CheckParsedIP(net.IP) (bool, error) {
	m.checkParsedCalls++
	if m.checkParsedErr != nil {
		return false, m.checkParsedErr
	}
	return m.parsedMatch, nil
}

func (m *parsedPathMockProvider) GetName() string {
	return m.name
}

func TestCheckCloudIP_UsesParsedProviderPath(t *testing.T) {
	mockProvider := &parsedPathMockProvider{
		name:        "AWS",
		parsedMatch: true,
	}

	checker := NewIPChecker(
		map[common.CloudProvider]provider.CloudProvider{
			common.AWS: mockProvider,
		},
		DefaultProviderOrder,
	)

	got, err := checker.checkCloudIp("192.168.1.1")
	if err != nil {
		t.Fatalf("checkCloudIp returned unexpected error: %v", err)
	}

	if got != common.AWS {
		t.Fatalf("checkCloudIp returned %q, want %q", got, common.AWS)
	}

	if mockProvider.initializeCalls != 1 {
		t.Fatalf("Initialize called %d times, want 1", mockProvider.initializeCalls)
	}

	if mockProvider.checkParsedCalls != 1 {
		t.Fatalf("CheckParsedIP called %d times, want 1", mockProvider.checkParsedCalls)
	}
}

func TestCheckCloudIP_InvalidIPShortCircuitsBeforeProviders(t *testing.T) {
	mockProvider := &parsedPathMockProvider{
		name:        "AWS",
		parsedMatch: true,
	}

	checker := NewIPChecker(
		map[common.CloudProvider]provider.CloudProvider{
			common.AWS: mockProvider,
		},
		DefaultProviderOrder,
	)

	_, err := checker.checkCloudIp("not-an-ip")
	if err == nil {
		t.Fatal("expected invalid IP to return an error")
	}

	if mockProvider.initializeCalls != 0 {
		t.Fatalf("Initialize called %d times for invalid IP, want 0", mockProvider.initializeCalls)
	}

	if mockProvider.checkParsedCalls != 0 {
		t.Fatalf("CheckParsedIP called %d times for invalid IP, want 0", mockProvider.checkParsedCalls)
	}
}
