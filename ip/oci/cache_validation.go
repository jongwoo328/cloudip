package oci

import (
	"cloudip/util"
	"errors"
	"fmt"
)

func (data *IpRangeDataOCI) validate() error {
	if data.LastUpdatedTimestamp == "" || len(data.Regions) == 0 {
		return errors.New("OCI data must contain a timestamp and IP ranges")
	}
	count := 0
	for _, region := range data.Regions {
		if region.Region == "" {
			return errors.New("OCI region name is missing")
		}
		for _, entry := range region.CIDRs {
			if err := util.ValidateCIDR(entry.CIDR, 32); err != nil {
				return fmt.Errorf("OCI %s: %w", region.Region, err)
			}
			count++
		}
		for _, entry := range region.IPv6CIDRs {
			if err := util.ValidateCIDR(entry.CIDR, 128); err != nil {
				return fmt.Errorf("OCI %s: %w", region.Region, err)
			}
			count++
		}
	}
	if count == 0 {
		return errors.New("OCI data must contain IP ranges")
	}
	return nil
}
