package digitalocean

import "cloudip/ip/provider"

type DigitalOceanProvider struct{ *provider.BaseProvider }

func NewDigitalOceanProvider() *DigitalOceanProvider {
	return newDigitalOceanProvider(ipDataManagerDigitalOcean)
}

func newDigitalOceanProvider(manager *IpDataManagerDigitalOcean) *DigitalOceanProvider {
	return &DigitalOceanProvider{BaseProvider: provider.NewBaseProvider("DigitalOcean", manager, func(bp *provider.BaseProvider) error {
		data, err := manager.LoadIpData()
		if err != nil {
			return err
		}
		for _, cidr := range data.CIDRs {
			if err := bp.AddCIDRRange(cidr); err != nil {
				return err
			}
		}
		return nil
	})}
}

var Provider = NewDigitalOceanProvider()
