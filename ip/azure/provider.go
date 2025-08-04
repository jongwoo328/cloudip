package azure

import (
	"cloudip/ip/provider"
	"cloudip/util"
	"fmt"
)

type AzureProvider struct {
	*provider.BaseProvider
}

func NewAzureProvider() *AzureProvider {
	return &AzureProvider{
		BaseProvider: provider.NewBaseProvider("Azure", ipDataManagerAzure),
	}
}

func (azure *AzureProvider) Initialize() error {
	err := azure.BaseProvider.Initialize()
	if err != nil {
		return err
	}

	azureIpRangeData := *ipDataManagerAzure.LoadIpData()

	for _, dataObject := range azureIpRangeData.Values {
		for _, prefix := range dataObject.Properties.AddressPrefixes {
			err := azure.AddCIDRRange(prefix)
			if err != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(err, fmt.Sprintf("Error parsing CIDR: %s", prefix)))
				continue
			}
		}
	}

	return nil
}

func (ipDataManager *IpDataManagerAzure) GetDataURL() string {
	return ipDataManager.DataURI
}

var Provider = NewAzureProvider()

func Initialize() {
	Provider.Initialize()
}

func IsAzureIp(ip string) (bool, error) {
	return Provider.CheckIP(ip)
}