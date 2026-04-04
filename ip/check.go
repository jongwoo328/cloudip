package ip

import (
	"cloudip/common"
	"cloudip/ip/aws"
	"cloudip/ip/azure"
	"cloudip/ip/gcp"
	"cloudip/ip/provider"
)

func Check(ips []string) []common.Result {
	results := make([]common.Result, len(ips))

	for index, ip := range ips {
		provider, err := checkCloudIp(ip)
		results[index] = common.Result{
			Ip:       ip,
			Provider: provider,
			Error:    err,
		}
	}

	return results
}

var cloudProviders = map[common.CloudProvider]provider.CloudProvider{
	common.AWS:   aws.Provider,
	common.GCP:   gcp.Provider,
	common.Azure: azure.Provider,
}

func checkCloudIp(ip string) (common.CloudProvider, error) {
	for _, providerType := range Providers {
		p, exists := cloudProviders[providerType]
		if !exists {
			continue
		}

		err := p.Initialize()
		if err != nil {
			return "", err
		}

		isMatch, err := p.CheckIP(ip)
		if err != nil {
			return "", err
		}

		if isMatch {
			return providerType, nil
		}
	}
	return "", nil
}
