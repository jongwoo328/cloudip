package digitalocean

import (
	"cloudip/util"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// The official feed uses RFC 8805's five columns. Geographic fields are not
// needed for provider identification; CSV parsing still validates the whole feed.
func decodeGeofeed(reader io.Reader) (*IpRangeDataDigitalOcean, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comment = '#'
	csvReader.FieldsPerRecord = 5
	data := &IpRangeDataDigitalOcean{}
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid DigitalOcean geofeed CSV: %w", err)
		}
		data.CIDRs = append(data.CIDRs, strings.TrimSpace(record[0]))
	}
	if err := data.validate(); err != nil {
		return nil, err
	}
	return data, nil
}

func (data *IpRangeDataDigitalOcean) validate() error {
	if len(data.CIDRs) == 0 {
		return errors.New("DigitalOcean data must contain IP ranges")
	}
	for _, cidr := range data.CIDRs {
		if err := util.ValidateCIDR(cidr, 0); err != nil {
			return err
		}
	}
	return nil
}

// Content signatures do not depend on optional or changing HTTP date headers.
func (data *IpRangeDataDigitalOcean) signature() string {
	digest := sha256.Sum256([]byte(strings.Join(data.CIDRs, "\n")))
	return hex.EncodeToString(digest[:])
}
