package gcp

import (
	"sync"

	"github.com/ip-api/cloudip/ip"
	"github.com/ip-api/cloudip/util"
)

var gcpChecker = ip.NewCloudIpChecker()
var initOnceGcp sync.Once

func Initialize() {
	initOnceGcp.Do(func() {
		err := ipDataManagerGcp.EnsureDataFile()
		if err != nil {
			util.PrintErrorTrace(err)
			return
		}

		gcpIpRangeData := ipDataManagerGcp.LoadIpData()
		if gcpIpRangeData == nil {
			// Handle case where data loading fails
			return
		}

		var ipv4Ranges []string
		var ipv6Ranges []string

		for _, prefix := range gcpIpRangeData.Prefixes {
			if prefix.Ipv4Prefix != "" {
				ipv4Ranges = append(ipv4Ranges, prefix.Ipv4Prefix)
			}
			// GCP data can have both IPv4 and IPv6 prefixes for the same entry,
			// so we don't use else if here.
			if prefix.Ipv6Prefix != "" {
				ipv6Ranges = append(ipv6Ranges, prefix.Ipv6Prefix)
			}
		}
		gcpChecker.Initialize(ipv4Ranges, ipv6Ranges)
	})
}

func IsGcpIp(ipAddr string) (bool, error) {
	return gcpChecker.IsCloudIp(ipAddr)
}
