package gcp

import (
	"cloudip/ip/provider"
)

type GCPProvider struct {
	*provider.BaseProvider
}

func NewGCPProvider() *GCPProvider {
	return &GCPProvider{
		BaseProvider: provider.NewBaseProvider("GCP", ipDataManagerGcp, func(bp *provider.BaseProvider) error {
			gcpIpRangeData := *ipDataManagerGcp.LoadIpData()

			for _, prefix := range gcpIpRangeData.Prefixes {
				if prefix.Ipv4Prefix != "" {
					bp.AddIPv4Range(prefix.Ipv4Prefix)
				} else if prefix.Ipv6Prefix != "" {
					bp.AddIPv6Range(prefix.Ipv6Prefix)
				}
			}

			return nil
		}),
	}
}

func (ipDataManager *IpDataManagerGcp) GetDataURL() string {
	return ipDataManager.DataURI
}

var Provider = NewGCPProvider()
