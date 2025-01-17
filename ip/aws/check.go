package aws

import (
	"cloudip/util"
	"fmt"
	"net"
	"sync"
)

var v4Tree *util.CIDRTree
var v6Tree *util.CIDRTree

var initialized = false
var initializeLock = sync.Mutex{}

func Initialize() {
	if initialized {
		return
	}

	initializeLock.Lock()
	defer initializeLock.Unlock()

	v4Tree = util.NewCIDRTree()
	v6Tree = util.NewCIDRTree()

	err := ipDataManagerAws.EnsureDataFile()
	if err != nil {
		util.PrintErrorTrace(err)
		return
	}

	awsIpRangeData := *ipDataManagerAws.LoadIpData()

	for _, prefix := range awsIpRangeData.Prefixes {
		v4Tree.AddCIDR(prefix.IpPrefix)
	}

	for _, prefix := range awsIpRangeData.Ipv6Prefixes {
		v6Tree.AddCIDR(prefix.Ipv6Prefix)
	}

	initialized = true
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
