package ip

import (
	"cloudip/common"
	"cloudip/ip/provider"
	"cloudip/util"
	"errors"
	"testing"
)

type mockProvider struct {
	name        string
	shouldMatch bool
	shouldError bool
	initError   bool
	initialized bool
	v4Tree      *util.CIDRTree
	v6Tree      *util.CIDRTree
}

func newMockProvider(name string, shouldMatch, shouldError, initError bool) *mockProvider {
	return &mockProvider{
		name:        name,
		shouldMatch: shouldMatch,
		shouldError: shouldError,
		initError:   initError,
		v4Tree:      util.NewCIDRTree(),
		v6Tree:      util.NewCIDRTree(),
	}
}

func (m *mockProvider) Initialize() error {
	if m.initError {
		return errors.New("initialization failed")
	}
	m.initialized = true

	// Add some test CIDR ranges
	if m.shouldMatch {
		m.v4Tree.AddCIDR("192.168.1.0/24")
		m.v4Tree.AddCIDR("10.0.0.0/8")
		m.v6Tree.AddCIDR("2001:db8::/32")
	}
	return nil
}

func (m *mockProvider) CheckIP(ip string) (bool, error) {
	if m.shouldError {
		return false, errors.New("check IP failed")
	}
	if !m.initialized {
		return false, errors.New("provider not initialized")
	}

	// Simple mock logic - just check if it should match
	if m.shouldMatch {
		// For testing purposes, match specific IPs
		switch ip {
		case "192.168.1.1", "10.1.1.1", "2001:db8::1":
			return true, nil
		}
	}
	return false, nil
}

func (m *mockProvider) GetName() string {
	return m.name
}

func TestCheckerCheck(t *testing.T) {
	tests := []struct {
		name             string
		ip               string
		mockProviders    map[common.CloudProvider]provider.CloudProvider
		expectedProvider common.CloudProvider
		expectError      bool
	}{
		{
			name: "AWS IP match",
			ip:   "192.168.1.1",
			mockProviders: map[common.CloudProvider]provider.CloudProvider{
				common.AWS: newMockProvider("AWS", true, false, false),
			},
			expectedProvider: common.AWS,
		},
		{
			name: "GCP IP match",
			ip:   "10.1.1.1",
			mockProviders: map[common.CloudProvider]provider.CloudProvider{
				common.AWS: newMockProvider("AWS", false, false, false),
				common.GCP: newMockProvider("GCP", true, false, false),
			},
			expectedProvider: common.GCP,
		},
		{
			name: "Azure IP match",
			ip:   "2001:db8::1",
			mockProviders: map[common.CloudProvider]provider.CloudProvider{
				common.AWS:   newMockProvider("AWS", false, false, false),
				common.GCP:   newMockProvider("GCP", false, false, false),
				common.Azure: newMockProvider("Azure", true, false, false),
			},
			expectedProvider: common.Azure,
		},
		{
			name: "No match found",
			ip:   "172.16.1.1",
			mockProviders: map[common.CloudProvider]provider.CloudProvider{
				common.AWS: newMockProvider("AWS", false, false, false),
				common.GCP: newMockProvider("GCP", false, false, false),
			},
			expectedProvider: "",
		},
		{
			name: "Initialization error",
			ip:   "192.168.1.1",
			mockProviders: map[common.CloudProvider]provider.CloudProvider{
				common.AWS: newMockProvider("AWS", true, false, true),
			},
			expectError: true,
		},
		{
			name: "CheckIP error",
			ip:   "192.168.1.1",
			mockProviders: map[common.CloudProvider]provider.CloudProvider{
				common.AWS: newMockProvider("AWS", true, true, false),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewIPChecker(tt.mockProviders, DefaultProviderOrder)
			results := checker.Check([]string{tt.ip})

			if len(results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results))
			}

			result := results[0]

			if tt.expectError {
				if result.Error == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if result.Error != nil {
				t.Errorf("Unexpected error: %v", result.Error)
				return
			}

			if result.Provider != tt.expectedProvider {
				t.Errorf("Provider mismatch: got %q, expected %q", result.Provider, tt.expectedProvider)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	checker := NewIPChecker(
		map[common.CloudProvider]provider.CloudProvider{
			common.AWS: newMockProvider("AWS", true, false, false),
			common.GCP: newMockProvider("GCP", false, false, false),
		},
		DefaultProviderOrder,
	)

	tests := []struct {
		name        string
		ips         []string
		expectedLen int
	}{
		{
			name:        "Single IP",
			ips:         []string{"192.168.1.1"},
			expectedLen: 1,
		},
		{
			name:        "Multiple IPs",
			ips:         []string{"192.168.1.1", "10.1.1.1", "172.16.1.1"},
			expectedLen: 3,
		},
		{
			name:        "Empty slice",
			ips:         []string{},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := checker.Check(tt.ips)

			if len(results) != tt.expectedLen {
				t.Errorf("Expected %d results, got %d", tt.expectedLen, len(results))
				return
			}

			for i, result := range results {
				if result.Ip != tt.ips[i] {
					t.Errorf("IP mismatch at index %d: got %s, expected %s", i, result.Ip, tt.ips[i])
				}
			}
		})
	}
}

func TestCheckWithProviderOrder(t *testing.T) {
	// Test that first matching provider wins
	checker := NewIPChecker(
		map[common.CloudProvider]provider.CloudProvider{
			common.AWS:   newMockProvider("AWS", true, false, false),
			common.GCP:   newMockProvider("GCP", true, false, false),
			common.Azure: newMockProvider("Azure", true, false, false),
		},
		DefaultProviderOrder,
	)

	ips := []string{"192.168.1.1"}
	results := checker.Check(ips)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Should only match AWS (first in provider order)
	if results[0].Provider != common.AWS {
		t.Errorf("Expected AWS, got %q", results[0].Provider)
	}
}
