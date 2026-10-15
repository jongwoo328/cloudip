package azure

import (
	"cloudip/util"
	"errors"
)

func (data *IpRangeDataAzure) validate() error {
	if data.ChangeNumber <= 0 {
		return errors.New("Azure data must contain a changeNumber")
	}
	count := 0
	for _, value := range data.Values {
		for _, cidr := range value.Properties.AddressPrefixes {
			if err := util.ValidateCIDR(cidr, 0); err != nil {
				return err
			}
			count++
		}
	}
	if count == 0 {
		return errors.New("Azure data must contain IP ranges")
	}
	return nil
}
