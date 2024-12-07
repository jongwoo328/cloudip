package aws

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

	metadataManager := GetMetadataManager()
	err := metadataManager.EnsureMetadataFile()
	if err != nil {
		util.PrintErrorTrace(err)
		return
	}

	ipDataManagerAws := GetIpDataManagerAws()
	err = ipDataManagerAws.EnsureDataFile()
	if err != nil {
		util.PrintErrorTrace(err)
		return
	}

	awsIpRangeData := IpRangeData{}

	for _, prefix := range awsIpRangeData.Prefixes {
		v4Tree.AddCIDR(prefix.IpPrefix)
	}

	for _, prefix := range awsIpRangeData.Ipv6Prefixes {
		v6Tree.AddCIDR(prefix.Ipv6Prefix)
	}
}

func IsAwsIp(ip string) (bool, error) {
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

func loadIpData() {

}
