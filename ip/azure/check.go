package azure

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

	err := ipDataManagerAzure.EnsureDataFile()
	if err != nil {
		util.PrintErrorTrace(err)
		return
	}

	azureIpRangeData := *ipDataManagerAzure.LoadIpData()

	for _, dataObject := range azureIpRangeData.Values {
		for _, prefix := range dataObject.Properties.AddressPrefixes {
			cidrVersion, err := util.GetCIDRVersion(prefix)
			if err != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(err, fmt.Sprintf("Error parsing CIDR: %s", prefix)))
				continue
			}

			if cidrVersion == util.IPv4 {
				v4Tree.AddCIDR(prefix)
			} else {
				v6Tree.AddCIDR(prefix)
			}
		}
	}

	initialized = true
}

func IsAzureIp(ip string) (bool, error) {
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
