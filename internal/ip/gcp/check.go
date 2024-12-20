package gcp

import (
	"cloudip/internal/util"
	"fmt"
	"net"
)

var v4Tree *util.CIDRTree
var v6Tree *util.CIDRTree

func init() {
	v4Tree = util.NewCIDRTree()
	v6Tree = util.NewCIDRTree()

	err := ensureMetadataFile()
	if err != nil {
		util.PrintErrorTrace(err)
		return
	}

	err = ipDataManagerGcp.EnsureDataFile()
	if err != nil {
		util.PrintErrorTrace(err)
		return
	}

	gcpIpRangeData := *ipDataManagerGcp.LoadIpData()

	for _, prefix := range gcpIpRangeData.Prefixes {
		if prefix.Ipv4Prefix != "" {
			v4Tree.AddCIDR(prefix.Ipv4Prefix)
		} else if prefix.Ipv6Prefix != "" {
			v6Tree.AddCIDR(prefix.Ipv6Prefix)
		}
	}

}

func IsGcpIp(ip string) (bool, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return false, fmt.Errorf("Error parsing IP: %s", ip)
	}

	if parsedIp.To4() != nil {
		return v4Tree.Match(ip), nil
	}
	if parsedIp.To16() != nil {
		return v6Tree.Match(ip), nil
	}
	return false, nil
}
