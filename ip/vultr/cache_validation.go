package vultr

import (
	"cloudip/util"
	"errors"
)

func (data *IpRangeDataVultr) validate() error {
	if data.Updated == "" || len(data.Subnets) == 0 {
		return errors.New("Vultr data must contain an update timestamp and IP ranges")
	}
	for _, subnet := range data.Subnets {
		if err := util.ValidateCIDR(subnet.IPPrefix, 0); err != nil {
			return err
		}
	}
	return nil
}
