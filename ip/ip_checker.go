package ip

import (
	"net"
	"sync"

	"github.com/ip-api/cloudip/util"
)

// CloudIpChecker holds CIDR trees for IPv4 and IPv6 addresses.
type CloudIpChecker struct {
	ipv4Tree *util.CIDRTree
	ipv6Tree *util.CIDRTree
	initOnce sync.Once
}

// NewCloudIpChecker initializes and returns a new CloudIpChecker.
func NewCloudIpChecker() *CloudIpChecker {
	return &CloudIpChecker{
		ipv4Tree: util.NewCIDRTree(),
		ipv6Tree: util.NewCIDRTree(),
	}
}

// Initialize populates the CIDR trees with the provided IP ranges.
// It ensures that initialization happens only once.
func (c *CloudIpChecker) Initialize(ipv4Ranges []string, ipv6Ranges []string) {
	c.initOnce.Do(func() {
		for _, cidr := range ipv4Ranges {
			c.ipv4Tree.AddCIDR(cidr)
		}
		for _, cidr := range ipv6Ranges {
			c.ipv6Tree.AddCIDR(cidr)
		}
	})
}

// IsCloudIp checks if the given IP string belongs to a cloud provider.
// It returns true if the IP is found in the CIDR trees, false otherwise,
// and an error if the IP string is invalid.
func (c *CloudIpChecker) IsCloudIp(ipStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, &net.ParseError{Type: "IP address", Text: ipStr}
	}

	if ip.To4() != nil {
		return c.ipv4Tree.Contains(ip), nil
	}
	return c.ipv6Tree.Contains(ip), nil
}
