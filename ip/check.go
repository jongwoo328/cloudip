package ip

import (
	"cloudip/common"
	"cloudip/ip/provider"
	"errors"
	"fmt"
	"net"
)

type IPChecker struct {
	providers     map[common.CloudProvider]provider.CloudProvider
	providerOrder []common.CloudProvider
	updatePolicy  common.UpdatePolicy
}

func NewIPChecker(providers map[common.CloudProvider]provider.CloudProvider, order []common.CloudProvider) *IPChecker {
	return &IPChecker{
		providers:     providers,
		providerOrder: order,
		updatePolicy:  common.DefaultUpdatePolicy(),
	}
}

func (c *IPChecker) SetUpdatePolicy(policy common.UpdatePolicy) {
	c.updatePolicy = policy
	for _, p := range c.providers {
		if setter, ok := p.(provider.UpdatePolicySetter); ok {
			setter.SetUpdatePolicy(policy)
		}
	}
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
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("error parsing IP: %s", ip)
	}

	var providerErr error
	for _, providerType := range c.providerOrder {
		p, exists := c.providers[providerType]
		if !exists {
			continue
		}

		err := p.Initialize()
		if err != nil {
			providerErr = errors.Join(providerErr, fmt.Errorf("%s initialize: %w", providerType, err))
			continue
		}

		isMatch, err := p.CheckParsedIP(parsedIP)
		if err != nil {
			providerErr = errors.Join(providerErr, fmt.Errorf("%s check: %w", providerType, err))
			continue
		}

		if isMatch {
			return providerType, nil
		}
	}
	return "", providerErr
}
