package oci

import "cloudip/ip/provider"

type OCIProvider struct{ *provider.BaseProvider }

func NewOCIProvider() *OCIProvider {
	return newOCIProvider(ipDataManagerOCI)
}

func newOCIProvider(manager *IpDataManagerOCI) *OCIProvider {
	return &OCIProvider{BaseProvider: provider.NewBaseProvider("OCI", manager, func(bp *provider.BaseProvider) error {
		data, err := manager.LoadIpData()
		if err != nil {
			return err
		}
		for _, region := range data.Regions {
			for _, entry := range region.CIDRs {
				if err := bp.AddIPv4Range(entry.CIDR); err != nil {
					return err
				}
			}
			for _, entry := range region.IPv6CIDRs {
				if err := bp.AddIPv6Range(entry.CIDR); err != nil {
					return err
				}
			}
		}
		return nil
	})}
}

var Provider = NewOCIProvider()
