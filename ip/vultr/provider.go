package vultr

import "cloudip/ip/provider"

type VultrProvider struct{ *provider.BaseProvider }

func NewVultrProvider() *VultrProvider { return newVultrProvider(ipDataManagerVultr) }

func newVultrProvider(manager *IpDataManagerVultr) *VultrProvider {
	return &VultrProvider{BaseProvider: provider.NewBaseProvider("Vultr", manager, func(bp *provider.BaseProvider) error {
		data, err := manager.LoadIpData()
		if err != nil {
			return err
		}
		for _, subnet := range data.Subnets {
			if err := bp.AddCIDRRange(subnet.IPPrefix); err != nil {
				return err
			}
		}
		return nil
	})}
}

var Provider = NewVultrProvider()
