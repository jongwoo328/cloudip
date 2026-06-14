package aws

import (
	"cloudip/ip/provider"
	"cloudip/util"
)

type AWSProvider struct {
	*provider.BaseProvider
}

func NewAWSProvider() *AWSProvider {
	return &AWSProvider{
		BaseProvider: provider.NewBaseProvider("AWS", ipDataManagerAws, func(bp *provider.BaseProvider) error {
			awsIpRangeData, err := ipDataManagerAws.LoadIpData()
			if err != nil {
				return err
			}

			for _, prefix := range awsIpRangeData.Prefixes {
				if err := bp.AddIPv4Range(prefix.IpPrefix); err != nil {
					util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing CIDR: "+prefix.IpPrefix))
					continue
				}
			}

			for _, prefix := range awsIpRangeData.Ipv6Prefixes {
				if err := bp.AddIPv6Range(prefix.Ipv6Prefix); err != nil {
					util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing CIDR: "+prefix.Ipv6Prefix))
					continue
				}
			}

			return nil
		}),
	}
}

var Provider = NewAWSProvider()
