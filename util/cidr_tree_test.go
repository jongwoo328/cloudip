package util

import (
	"testing"
)

func TestAddCIDRAndMatch(t *testing.T) {
	tree := NewCIDRTree()

	// Add test CIDRs
	tree.AddCIDR("2600:1f13:a0d:a700::/56")
	tree.AddCIDR("192.168.1.0/24")
	tree.AddCIDR("4.145.74.52/30")

	// Test cases
	tests := []struct {
		ip          string
		expected    bool
		description string
	}{
		{"2600:1f13:a0d:a700:0:0:0:1", true, "IPv6 address within CIDR"},
		{"2600:1f13:a0d:a700::1", true, "IPv6 address within CIDR"},
		{"2600:1f13:b0d:a700::1", false, "IPv6 address outside CIDR"},
		{"192.168.1.100", true, "IPv4 address within CIDR"},
		{"192.168.2.1", false, "IPv4 address outside CIDR"},
		{"invalid-ip", false, "Invalid IP format"},
		{"::ffff:192.168.1.1", true, "IPv4-mapped IPv6 address within CIDR"},
		{"4.145.74.53", true, "Simple test"},
	}

	// Execute tests
	for _, test := range tests {
		result := tree.Match(test.ip)
		if result != test.expected {
			t.Errorf("Failed: %s - got %v, expected %v", test.description, result, test.expected)
		}
	}
}
