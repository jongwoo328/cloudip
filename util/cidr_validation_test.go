package util

import "testing"

func TestValidateCIDR(t *testing.T) {
	for _, tc := range []struct {
		cidr  string
		bits  int
		valid bool
	}{
		{"10.0.0.0/8", 32, true}, {"2001:db8::/32", 128, true}, {"::/0", 0, true},
		{"bad", 0, false}, {"10.0.0.0/33", 32, false}, {"10.0.0.0/8", 128, false},
		{"2001:db8::/32", 32, false}, {"::ffff:192.0.2.0/120", 0, false},
	} {
		if err := ValidateCIDR(tc.cidr, tc.bits); (err == nil) != tc.valid {
			t.Errorf("ValidateCIDR(%q,%d)=%v", tc.cidr, tc.bits, err)
		}
	}
}
