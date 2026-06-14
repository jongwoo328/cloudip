package cloudflare

import (
	"cloudip/ip/provider"
	"cloudip/util"
)

type CloudflareProvider struct {
	*provider.BaseProvider
}

func NewCloudflareProvider() *CloudflareProvider {
	return &CloudflareProvider{
		BaseProvider: provider.NewBaseProvider("Cloudflare", ipDataManagerCloudflare, func(bp *provider.BaseProvider) error {
			data, err := ipDataManagerCloudflare.LoadIpData()
			if err != nil {
				return err
			}

			for _, cidr := range data.V4CIDRs {
				if err := bp.AddIPv4Range(cidr); err != nil {
					util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing CIDR: "+cidr))
					continue
				}
			}

			for _, cidr := range data.V6CIDRs {
				if err := bp.AddIPv6Range(cidr); err != nil {
					util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing CIDR: "+cidr))
					continue
				}
			}

			return nil
		}),
	}
}

var Provider = NewCloudflareProvider()
