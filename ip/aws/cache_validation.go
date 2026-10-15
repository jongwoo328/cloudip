package aws

import (
	"cloudip/util"
	"errors"
)

func (data *IpRangeDataAws) validate() error {
	if data.SyncToken == "" || len(data.Prefixes)+len(data.Ipv6Prefixes) == 0 {
		return errors.New("AWS data must contain a syncToken and IP ranges")
	}
	for _, prefix := range data.Prefixes {
		if err := util.ValidateCIDR(prefix.IpPrefix, 32); err != nil {
			return err
		}
	}
	for _, prefix := range data.Ipv6Prefixes {
		if err := util.ValidateCIDR(prefix.Ipv6Prefix, 128); err != nil {
			return err
		}
	}
	return nil
}
