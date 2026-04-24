package cloudflare

import (
	"cloudip/ip/provider"
)

type CloudflareProvider struct {
	*provider.BaseProvider
}

func NewCloudflareProvider() *CloudflareProvider {
	return &CloudflareProvider{
		BaseProvider: provider.NewBaseProvider("Cloudflare", ipDataManagerCloudflare, func(bp *provider.BaseProvider) error {
			data := *ipDataManagerCloudflare.LoadIpData()

			for _, cidr := range data.V4CIDRs {
				bp.AddIPv4Range(cidr)
			}

			for _, cidr := range data.V6CIDRs {
				bp.AddIPv6Range(cidr)
			}

			return nil
		}),
	}
}

var Provider = NewCloudflareProvider()
