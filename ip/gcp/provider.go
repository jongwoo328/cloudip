package gcp

import (
	"cloudip/ip/provider"
	"cloudip/util"
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

func Initialize() error {
	err := Provider.Initialize()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "GCP provider initialization failed"))
	}
	return err
}

func IsGcpIp(ip string) (bool, error) {
	return Provider.CheckIP(ip)
}