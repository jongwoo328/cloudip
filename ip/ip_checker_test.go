package ip

import (
	"testing"
)

func TestNewCloudIpChecker(t *testing.T) {
	checker := NewCloudIpChecker()
	if checker == nil {
		t.Fatal("NewCloudIpChecker() returned nil")
	}
	if checker.ipv4Tree == nil {
		t.Error("NewCloudIpChecker().ipv4Tree is nil")
	}
	if checker.ipv6Tree == nil {
		t.Error("NewCloudIpChecker().ipv6Tree is nil")
	}
}

func TestInitializeAndIsCloudIp(t *testing.T) {
	checker := NewCloudIpChecker()

	initialIPv4Ranges := []string{"10.0.0.0/8", "172.16.0.0/12"}
	initialIPv6Ranges := []string{"2001:db8::/32"}

	checker.Initialize(initialIPv4Ranges, initialIPv6Ranges)

	testCases := []struct {
		name        string
		ip          string
		expectMatch bool
		expectError bool
	}{
		{"IPv4 in initial range 1", "10.1.2.3", true, false},
		{"IPv4 in initial range 2", "172.20.0.1", true, false},
		{"IPv4 not in initial range", "1.1.1.1", false, false},
		{"IPv6 in initial range", "2001:db8:1:2::1", true, false},
		{"IPv6 not in initial range", "2002::1", false, false},
		{"Invalid IP", "not-an-ip", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_initial_load", func(t *testing.T) {
			match, err := checker.IsCloudIp(tc.ip)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for IP %s, got nil", tc.ip)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for IP %s: %v", tc.ip, err)
				}
			}
			if match != tc.expectMatch {
				t.Errorf("Expected match %v for IP %s, got %v", tc.expectMatch, tc.ip, match)
			}
		})
	}

	// Test sync.Once: try to initialize again with different ranges
	differentIPv4Ranges := []string{"192.168.0.0/16"}
	differentIPv6Ranges := []string{"2003:cafe::/32"}
	checker.Initialize(differentIPv4Ranges, differentIPv6Ranges)

	// Re-run tests; they should still match against the *initial* ranges
	t.Run("sync_once_check", func(t *testing.T) {
		for _, tc := range testCases {
			t.Run(tc.name+"_after_second_init_attempt", func(t *testing.T) {
				match, err := checker.IsCloudIp(tc.ip)
				if tc.expectError {
					if err == nil {
						t.Errorf("Expected error for IP %s, got nil (after 2nd init)", tc.ip)
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error for IP %s: %v (after 2nd init)", tc.ip, err)
					}
				}
				if match != tc.expectMatch {
					t.Errorf("Expected match %v for IP %s, got %v (after 2nd init - sync.Once failed?)", tc.expectMatch, tc.ip, match)
				}
			})
		}
		// Explicitly check if an IP from the *different* range matches (it shouldn't)
		match, _ := checker.IsCloudIp("192.168.1.1")
		if match {
			t.Errorf("IP 192.168.1.1 from second init attempt matched, sync.Once failed")
		}
		match, _ = checker.IsCloudIp("2003:cafe::1")
		if match {
			t.Errorf("IP 2003:cafe::1 from second init attempt matched, sync.Once failed")
		}
	})
}

func TestInitializeEmptyRanges(t *testing.T) {
	checker := NewCloudIpChecker()
	checker.Initialize([]string{}, []string{})

	testIPs := []string{"10.1.2.3", "2001:db8::1", "0.0.0.0"}
	for _, ip := range testIPs {
		t.Run(ip, func(t *testing.T) {
			match, err := checker.IsCloudIp(ip)
			if err != nil {
				t.Errorf("Unexpected error for IP %s: %v", ip, err)
			}
			if match {
				t.Errorf("Expected no match for IP %s with empty ranges, got match", ip)
			}
		})
	}
}

func TestInitializeOnlyV4(t *testing.T) {
	checker := NewCloudIpChecker()
	v4Ranges := []string{"192.0.2.0/24"}
	checker.Initialize(v4Ranges, []string{})

	testCases := []struct {
		name        string
		ip          string
		expectMatch bool
	}{
		{"IPv4 in range", "192.0.2.100", true},
		{"IPv4 not in range", "198.51.100.1", false},
		{"IPv6 any", "2001:db8::1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := checker.IsCloudIp(tc.ip)
			if err != nil {
				t.Errorf("Unexpected error for IP %s: %v", tc.ip, err)
			}
			if match != tc.expectMatch {
				t.Errorf("Expected match %v for IP %s, got %v", tc.expectMatch, tc.ip, match)
			}
		})
	}
}

func TestInitializeOnlyV6(t *testing.T) {
	checker := NewCloudIpChecker()
	v6Ranges := []string{"2001:db8:abcd::/48"}
	checker.Initialize([]string{}, v6Ranges)

	testCases := []struct {
		name        string
		ip          string
		expectMatch bool
	}{
		{"IPv6 in range", "2001:db8:abcd:0001::1234", true},
		{"IPv6 not in range", "2001:db8:ffff::1", false},
		{"IPv4 any", "192.0.2.1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := checker.IsCloudIp(tc.ip)
			if err != nil {
				t.Errorf("Unexpected error for IP %s: %v", tc.ip, err)
			}
			if match != tc.expectMatch {
				t.Errorf("Expected match %v for IP %s, got %v", tc.expectMatch, tc.ip, match)
			}
		})
	}
}

// Test case for an IPv4 address when only IPv6 ranges are loaded
func TestIsCloudIp_IPv4_OnlyIPv6Loaded(t *testing.T) {
	checker := NewCloudIpChecker()
	checker.Initialize([]string{}, []string{"2001:db8::/32"})
	match, err := checker.IsCloudIp("10.1.2.3")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if match {
		t.Error("Expected false for IPv4 check when only IPv6 ranges are loaded, got true")
	}
}

// Test case for an IPv6 address when only IPv4 ranges are loaded
func TestIsCloudIp_IPv6_OnlyIPv4Loaded(t *testing.T) {
	checker := NewCloudIpChecker()
	checker.Initialize([]string{"10.0.0.0/8"}, []string{})
	match, err := checker.IsCloudIp("2001:db8::1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if match {
		t.Error("Expected false for IPv6 check when only IPv4 ranges are loaded, got true")
	}
}

// Test case for Initialize being called multiple times (sync.Once specific detailed check)
// This is partially covered in TestInitializeAndIsCloudIp, but can be more direct.
func TestInitialize_SyncOnce_Direct(t *testing.T) {
	checker := NewCloudIpChecker()

	firstIPv4 := []string{"1.1.1.0/24"}
	firstIPv6 := []string{"2001::/120"}
	checker.Initialize(firstIPv4, firstIPv6)

	// Check if an IP from the first set matches
	match, _ := checker.IsCloudIp("1.1.1.1")
	if !match {
		t.Error("IP from first init (1.1.1.1) did not match after first init")
	}
	match, _ = checker.IsCloudIp("2001::1")
	if !match {
		t.Error("IP from first init (2001::1) did not match after first init")
	}


	secondIPv4 := []string{"2.2.2.0/24"}
	secondIPv6 := []string{"2002::/120"}
	// Attempt to re-initialize. This should not change the underlying trees.
	checker.Initialize(secondIPv4, secondIPv6)

	// Check again if an IP from the first set matches (it should)
	match, _ = checker.IsCloudIp("1.1.1.1")
	if !match {
		t.Error("IP from first init (1.1.1.1) did not match after second init attempt")
	}
	match, _ = checker.IsCloudIp("2001::1")
	if !match {
		t.Error("IP from first init (2001::1) did not match after second init attempt")
	}

	// Check if an IP from the second set matches (it should NOT)
	match, _ = checker.IsCloudIp("2.2.2.1")
	if match {
		t.Error("IP from second init (2.2.2.1) matched, sync.Once failed")
	}
	match, _ = checker.IsCloudIp("2002::1")
	if match {
		t.Error("IP from second init (2002::1) matched, sync.Once failed")
	}
}

// Test for IsCloudIp with nil trees (should not happen with NewCloudIpChecker, but defensive)
// This test might be of limited value as NewCloudIpChecker ensures non-nil trees.
// And Initialize would panic if trees were nil and it tried to add.
// However, if the internal structure of CloudIpChecker changed, this might be relevant.
// For now, this is more of a conceptual test.
func TestIsCloudIp_NilTrees(t *testing.T) {
	checker := &CloudIpChecker{} // Intentionally creating a checker without NewCloudIpChecker
	// This will cause a panic if IsCloudIp tries to access methods on nil trees.
	// A more robust test would check if IsCloudIp handles this gracefully,
	// but current implementation would panic, which is a valid failure.
	// We expect IsCloudIp to not panic if trees are nil, but rather return false.
	// However, the current util.CIDRTree.Contains() would panic.
	// So, this test verifies current behavior or highlights a potential fragility
	// if trees could somehow become nil post-initialization (which they shouldn't).

	// Let's ensure our test setup itself doesn't panic immediately
	if checker.ipv4Tree != nil || checker.ipv6Tree != nil {
		t.Log("Skipping direct nil tree test as trees are somehow initialized or this test needs rethink")
		return
	}
	
	// Test with an IP. If Contains panics on nil receiver, this test will fail.
	// This implicitly tests that IsCloudIp doesn't proceed if trees are nil,
	// or that Contains() on a nil tree (if that were possible) returns false.
	// Given current CIDRTree, Contains() on nil tree would panic.
	// The CloudIpChecker.IsCloudIp itself does not check for nil trees before calling Contains.
	// This is acceptable because NewCloudIpChecker guarantees non-nil trees.
	
	// A "safer" version of this test:
	checkerWithNilTree := NewCloudIpChecker() // Start with a valid one
	checkerWithNilTree.ipv4Tree = nil          // Force a nil tree (only possible in same package)
	
	match, err := checkerWithNilTree.IsCloudIp("1.2.3.4")
	if err != nil {
		t.Errorf("IsCloudIp with one nil tree returned error: %v", err)
	}
	if match {
		t.Errorf("IsCloudIp with one nil tree returned match true, expected false")
	}

	checkerWithNilTree.ipv4Tree = checkerWithNilTree.ipv6Tree // restore for next part
	checkerWithNilTree.ipv6Tree = nil 
	match, err = checkerWithNilTree.IsCloudIp("2001:db8::1")
	if err != nil {
		t.Errorf("IsCloudIp with other nil tree returned error: %v", err)
	}
	if match {
		t.Errorf("IsCloudIp with other nil tree returned match true, expected false")
	}
}
