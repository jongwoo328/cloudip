package gcp

import (
	"cloudip/util"
	"errors"
)

func (data *IpRangeDataGcp) validate() error {
	if data.SyncToken == "" || len(data.Prefixes) == 0 {
		return errors.New("GCP data must contain a syncToken and IP ranges")
	}
	for _, prefix := range data.Prefixes {
		if (prefix.Ipv4Prefix == "") == (prefix.Ipv6Prefix == "") {
			return errors.New("GCP prefix must contain exactly one IP range")
		}
		if prefix.Ipv4Prefix != "" {
			if err := util.ValidateCIDR(prefix.Ipv4Prefix, 32); err != nil {
				return err
			}
		} else {
			if err := util.ValidateCIDR(prefix.Ipv6Prefix, 128); err != nil {
				return err
			}
		}
	}
	return nil
}
