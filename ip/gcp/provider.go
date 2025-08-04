package gcp

import (
	"cloudip/ip/provider"
)

type GCPProvider struct {
	*provider.BaseProvider
}

func NewGCPProvider() *GCPProvider {
	return &GCPProvider{
		BaseProvider: provider.NewBaseProvider("GCP", ipDataManagerGcp),
	}
}

func (gcp *GCPProvider) Initialize() error {
	err := gcp.BaseProvider.Initialize()
	if err != nil {
		return err
	}

	gcpIpRangeData := *ipDataManagerGcp.LoadIpData()

	for _, prefix := range gcpIpRangeData.Prefixes {
		if prefix.Ipv4Prefix != "" {
			gcp.AddIPv4Range(prefix.Ipv4Prefix)
		} else if prefix.Ipv6Prefix != "" {
			gcp.AddIPv6Range(prefix.Ipv6Prefix)
		}
	}

	return nil
}

func (ipDataManager *IpDataManagerGcp) GetDataURL() string {
	return ipDataManager.DataURI
}

var Provider = NewGCPProvider()

func Initialize() {
	Provider.Initialize()
}

func IsGcpIp(ip string) (bool, error) {
	return Provider.CheckIP(ip)
}