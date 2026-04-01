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
			awsIpRangeData := *ipDataManagerAws.LoadIpData()

			for _, prefix := range awsIpRangeData.Prefixes {
				bp.AddIPv4Range(prefix.IpPrefix)
			}

			for _, prefix := range awsIpRangeData.Ipv6Prefixes {
				bp.AddIPv6Range(prefix.Ipv6Prefix)
			}

			return nil
		}),
	}
}

func (ipDataManager *IpDataManagerAws) GetDataURL() string {
	return ipDataManager.DataURI
}

var Provider = NewAWSProvider()

func Initialize() error {
	err := Provider.Initialize()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "AWS provider initialization failed"))
	}
	return err
}

func IsAwsIp(ip string) (bool, error) {
	return Provider.CheckIP(ip)
}
