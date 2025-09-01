package ip

import (
	"cloudip/common"
	"cloudip/ip/aws"
	"cloudip/ip/azure"
	"cloudip/ip/gcp"
	"cloudip/ip/provider"
)

func CheckIp(ips *[]string) []common.CheckIpResult {
	results := make([]common.CheckIpResult, len(*ips))

	for index, ip := range *ips {
		checkResult, err := checkCloudIp(ip)
		results[index] = common.CheckIpResult{
			Ip:     ip,
			Result: checkResult,
			Error:  err,
		}
	}

	return results
}

var cloudProviders = map[common.CloudProvider]provider.CloudProvider{
	common.AWS:   aws.Provider,
	common.GCP:   gcp.Provider,
	common.Azure: azure.Provider,
}

func checkCloudIp(ip string) (common.Result, error) {
	result := common.Result{}

	for _, providerType := range Providers {
		provider, exists := cloudProviders[providerType]
		if !exists {
			continue
		}

		err := provider.Initialize()
		if err != nil {
			return result, err
		}

		isMatch, err := provider.CheckIP(ip)
		if err != nil {
			return result, err
		}

		if isMatch {
			switch providerType {
			case common.AWS:
				result.Aws = true
			case common.GCP:
				result.Gcp = true
			case common.Azure:
				result.Azure = true
			}
			return result, nil
		}
	}
	return result, nil
}
