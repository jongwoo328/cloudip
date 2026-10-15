package cloudflare

import (
	"cloudip/util"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (data *IpRangeDataCloudflare) validate() error {
	if data.IsEmpty() {
		return errors.New("Cloudflare data must contain IP ranges")
	}
	for _, cidr := range data.V4CIDRs {
		if err := util.ValidateCIDR(cidr, 32); err != nil {
			return err
		}
	}
	for _, cidr := range data.V6CIDRs {
		if err := util.ValidateCIDR(cidr, 128); err != nil {
			return err
		}
	}
	return nil
}

func (data *ipListResponseCloudflare) validate() error {
	if !data.Success {
		return errors.New("Cloudflare IP API returned unsuccessful response")
	}
	if _, err := data.signature(); err != nil {
		return err
	}
	ranges := IpRangeDataCloudflare{V4CIDRs: data.Result.V4CIDRs, V6CIDRs: data.Result.V6CIDRs}
	return ranges.validate()
}

// Both address families are committed by a single rename. Legacy text caches
// remain readable until an update creates the first JSON snapshot.
func (m *IpDataManagerCloudflare) snapshotPath() string {
	return filepath.Join(filepath.Dir(m.DataFilePathV4), "cloudflare.json")
}

func (m *IpDataManagerCloudflare) hasSnapshot() (bool, error) {
	info, err := os.Stat(m.snapshotPath())
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("expected cache file at %q", m.snapshotPath())
	}
	return true, nil
}

func (m *IpDataManagerCloudflare) readDataFile() (*IpRangeDataCloudflare, error) {
	file, err := os.Open(m.snapshotPath())
	if err == nil {
		defer file.Close()
		data := IpRangeDataCloudflare{}
		if err := util.ReadJSON(file, &data); err != nil {
			return nil, err
		}
		return &data, data.validate()
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	v4, err := readCIDRLines(m.DataFilePathV4)
	if err != nil {
		return nil, err
	}
	v6, err := readCIDRLines(m.DataFilePathV6)
	if err != nil {
		return nil, err
	}
	data := &IpRangeDataCloudflare{V4CIDRs: v4, V6CIDRs: v6}
	return data, data.validate()
}
