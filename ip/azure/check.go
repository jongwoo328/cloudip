package azure

import (
	"sync"
	// "fmt" // No longer needed

	"github.com/ip-api/cloudip/ip"
	"github.com/ip-api/cloudip/util"
)

var azureChecker = ip.NewCloudIpChecker()
var initOnceAzure sync.Once

func Initialize() {
	initOnceAzure.Do(func() {
		err := ipDataManagerAzure.EnsureDataFile()
		if err != nil {
			util.PrintErrorTrace(err)
			return
		}

		azureIpRangeData := ipDataManagerAzure.LoadIpData()
		if azureIpRangeData == nil {
			// Handle case where data loading fails, e.g. log an error
			return
		}

		var ipv4Ranges []string
		var ipv6Ranges []string

		for _, dataObject := range azureIpRangeData.Values {
			for _, prefix := range dataObject.Properties.AddressPrefixes {
				cidrVersion, err := util.GetCIDRVersion(prefix)
				if err != nil {
					// Consider logging this error more formally if needed
					util.PrintErrorTrace(util.ErrorWithInfo(err, "Error parsing Azure CIDR: "+prefix))
					continue
				}

				if cidrVersion == util.IPv4 {
					ipv4Ranges = append(ipv4Ranges, prefix)
				} else if cidrVersion == util.IPv6 {
					ipv6Ranges = append(ipv6Ranges, prefix)
				}
			}
		}
		azureChecker.Initialize(ipv4Ranges, ipv6Ranges)
	})
}

func IsAzureIp(ipAddr string) (bool, error) {
	return azureChecker.IsCloudIp(ipAddr)
}
