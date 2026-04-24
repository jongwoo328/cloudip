package cloudflare

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

type IpDataManagerCloudflare struct {
	DataURI        string
	DataFileV4     string
	DataFileV6     string
	DataFilePathV4 string
	DataFilePathV6 string
	IpRange        IpRangeDataCloudflare
}

type IpRangeDataCloudflare struct {
	V4CIDRs []string
	V6CIDRs []string
}

type ipListResponseCloudflare struct {
	Success bool `json:"success"`
	Result  struct {
		Etag    string   `json:"etag"`
		V4CIDRs []string `json:"ipv4_cidrs"`
		V6CIDRs []string `json:"ipv6_cidrs"`
	} `json:"result"`
}

func (ipRange IpRangeDataCloudflare) IsEmpty() bool {
	return len(ipRange.V4CIDRs) == 0 && len(ipRange.V6CIDRs) == 0
}

func (m *IpDataManagerCloudflare) GetSignatureUpstream() (string, error) {
	data, err := m.fetchData()
	if err != nil {
		return "", err
	}
	return data.signature()
}

func (m *IpDataManagerCloudflare) downloadData() error {
	common.VerboseOutput("Downloading Cloudflare IP ranges...")
	if m.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	data, err := m.fetchData()
	if err != nil {
		return err
	}
	return m.writeData(data)
}

func (m *IpDataManagerCloudflare) fetchData() (*ipListResponseCloudflare, error) {
	data := ipListResponseCloudflare{}
	_, err := util.DownloadJSONFromUrl(m.DataURI, &data)
	if err != nil {
		return nil, err
	}
	if !data.Success {
		return nil, errors.New("Cloudflare IP API returned unsuccessful response")
	}
	if _, err := data.signature(); err != nil {
		return nil, err
	}
	return &data, nil
}

func (m *IpDataManagerCloudflare) writeData(data *ipListResponseCloudflare) error {
	signature, err := data.signature()
	if err != nil {
		return err
	}

	if err := writeCIDRLines(m.DataFilePathV4, data.Result.V4CIDRs); err != nil {
		return err
	}
	if err := writeCIDRLines(m.DataFilePathV6, data.Result.V6CIDRs); err != nil {
		return err
	}

	if metadataManager.IsSignatureExpired(signature) {
		metadata := common.CloudMetadata{
			Type:      common.Cloudflare,
			Signature: signature,
		}
		if err := metadataManager.Write(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput("Cloudflare IP ranges updated")
	}

	return nil
}

func (m *IpDataManagerCloudflare) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	if !util.IsFileExists(m.DataFilePathV4) || !util.IsFileExists(m.DataFilePathV6) {
		common.VerboseOutput("Cloudflare IP ranges file not exists.")
		return m.downloadData()
	}
	if isExpired() {
		common.VerboseOutput("Cloudflare IP ranges are outdated. Updating to the latest version...")
		return m.downloadData()
	}
	common.VerboseOutput("Cloudflare IP ranges are up-to-date.")

	return nil
}

func (m *IpDataManagerCloudflare) LoadIpData() *IpRangeDataCloudflare {
	if !m.IpRange.IsEmpty() {
		return &m.IpRange
	}

	data := IpRangeDataCloudflare{
		V4CIDRs: readCIDRLines(m.DataFilePathV4),
		V6CIDRs: readCIDRLines(m.DataFilePathV6),
	}

	m.IpRange = data
	return &m.IpRange
}

func readCIDRLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		err = util.ErrorWithInfo(err, "error opening data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		err = util.ErrorWithInfo(err, "error reading data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}

	lines := strings.Split(string(content), "\n")
	cidrs := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		cidrs = append(cidrs, trimmed)
	}
	return cidrs
}

func writeCIDRLines(path string, cidrs []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		err = util.ErrorWithInfo(err, "error opening data file")
		util.PrintErrorTrace(err)
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(strings.Join(cidrs, "\n") + "\n"); err != nil {
		err = util.ErrorWithInfo(err, "error writing data file")
		util.PrintErrorTrace(err)
		return err
	}
	return nil
}

func (data *ipListResponseCloudflare) signature() (string, error) {
	if data.Result.Etag == "" {
		return "", errors.New("cannot get etag")
	}
	return data.Result.Etag, nil
}

var ipDataManagerCloudflare = &IpDataManagerCloudflare{
	DataURI:        getDataUrl(),
	DataFileV4:     DataFileV4,
	DataFileV6:     DataFileV6,
	DataFilePathV4: DataFilePathCloudflareV4,
	DataFilePathV6: DataFilePathCloudflareV6,
	IpRange:        IpRangeDataCloudflare{},
}
