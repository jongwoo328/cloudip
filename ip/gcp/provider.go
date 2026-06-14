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
			gcpIpRangeData, err := ipDataManagerGcp.LoadIpData()
			if err != nil {
				return err
			}

			for _, prefix := range gcpIpRangeData.Prefixes {
				if prefix.Ipv4Prefix != "" {
					if err := bp.AddIPv4Range(prefix.Ipv4Prefix); err != nil {
						util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing CIDR: "+prefix.Ipv4Prefix))
						continue
					}
				} else if prefix.Ipv6Prefix != "" {
					if err := bp.AddIPv6Range(prefix.Ipv6Prefix); err != nil {
						util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing CIDR: "+prefix.Ipv6Prefix))
						continue
					}
				}
			}

			return nil
		}),
	}
}

var Provider = NewGCPProvider()
