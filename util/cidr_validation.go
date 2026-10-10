package util

import (
	"fmt"
	"net"
)

// ValidateCIDR checks the address family as well as syntax before data replaces a
// usable cache. bits=0 accepts either family, but mapped IPv6 is unsupported by
// the separate IPv4/IPv6 trees.
func ValidateCIDR(cidr string, bits int) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}
	_, size := network.Mask.Size()
	if bits != 0 && size != bits {
		return fmt.Errorf("CIDR %q has address width %d, expected %d", cidr, size, bits)
	}
	if size == 128 && ip.To4() != nil {
		return fmt.Errorf("IPv4-mapped IPv6 CIDR %q is unsupported", cidr)
	}
	return nil
}
