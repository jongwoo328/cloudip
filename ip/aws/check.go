package aws

import (
	"sync"

	"github.com/ip-api/cloudip/ip"
	"github.com/ip-api/cloudip/util" // Keep for PrintErrorTrace, consider refactoring if not needed elsewhere
)

var awsChecker = ip.NewCloudIpChecker()
var initOnceAws sync.Once

func Initialize() {
	initOnceAws.Do(func() {
		err := ipDataManagerAws.EnsureDataFile()
		if err != nil {
			util.PrintErrorTrace(err) // Assuming PrintErrorTrace is still relevant
			return
		}

		awsIpRangeData := ipDataManagerAws.LoadIpData()
		if awsIpRangeData == nil {
			// Handle case where data loading fails, e.g. log an error
			// For now, assume LoadIpData handles its errors or returns non-nil
			return
		}

		var ipv4Ranges []string
		for _, prefix := range awsIpRangeData.Prefixes {
			ipv4Ranges = append(ipv4Ranges, prefix.IpPrefix)
		}

		var ipv6Ranges []string
		for _, prefix := range awsIpRangeData.Ipv6Prefixes {
			ipv6Ranges = append(ipv6Ranges, prefix.Ipv6Prefix)
		}

		awsChecker.Initialize(ipv4Ranges, ipv6Ranges)
	})
}

func IsAwsIp(ipAddr string) (bool, error) {
	return awsChecker.IsCloudIp(ipAddr)
}
