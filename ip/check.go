package ip

import (
	"cloudip/common"
	"cloudip/ip/provider"
)

type IPChecker struct {
	providers     map[common.CloudProvider]provider.CloudProvider
	providerOrder []common.CloudProvider
}

func NewIPChecker(providers map[common.CloudProvider]provider.CloudProvider, order []common.CloudProvider) *IPChecker {
	return &IPChecker{providers: providers, providerOrder: order}
}

func (c *IPChecker) Check(ips []string) []common.Result {
	results := make([]common.Result, len(ips))

	for index, ip := range ips {
		provider, err := c.checkCloudIp(ip)
		results[index] = common.Result{
			Ip:       ip,
			Provider: provider,
			Error:    err,
		}
	}

	return results
}

func (c *IPChecker) checkCloudIp(ip string) (common.CloudProvider, error) {
	for _, providerType := range c.providerOrder {
		p, exists := c.providers[providerType]
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
