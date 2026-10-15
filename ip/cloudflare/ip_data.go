package cloudflare

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

type IpDataManagerCloudflare struct {
	DataURI        string
	DataFileV4     string
	DataFileV6     string
	DataFilePathV4 string
	DataFilePathV6 string
	IpRange        IpRangeDataCloudflare
	UpdatePolicy   common.UpdatePolicy
}

type IpRangeDataCloudflare struct {
	V4CIDRs []string `json:"ipv4_cidrs"`
	V6CIDRs []string `json:"ipv6_cidrs"`
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
	if err := data.validate(); err != nil {
		return nil, err
	}
	return &data, nil
}

func (m *IpDataManagerCloudflare) writeData(data *ipListResponseCloudflare) error {
	signature, err := data.signature()
	if err != nil {
		return err
	}

	if err := data.validate(); err != nil {
		return err
	}
	ranges := IpRangeDataCloudflare{V4CIDRs: data.Result.V4CIDRs, V6CIDRs: data.Result.V6CIDRs}
	if err := util.WriteJSONFile(m.snapshotPath(), &ranges); err != nil {
		return err
	}

	signatureExpired := metadataManager.IsSignatureExpired(signature)
	metadata := common.CloudMetadata{
		Type:        common.Cloudflare,
		Signature:   signature,
		LastChecked: time.Now().Unix(),
	}
	if err := metadataManager.Write(&metadata); err != nil {
		err = util.ErrorWithInfo(err, "error writing metadata")
		util.PrintErrorTrace(err)
		return err
	}
	if signatureExpired {
		common.VerboseOutput("Cloudflare IP ranges updated")
	}

	return nil
}

func (m *IpDataManagerCloudflare) SetUpdatePolicy(policy common.UpdatePolicy) {
	m.UpdatePolicy = policy
}

func (m *IpDataManagerCloudflare) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	snapshotExists, err := m.hasSnapshot()
	if err != nil {
		return err
	}
	if snapshotExists {
		if err := metadataManager.Read(); err != nil {
			return err
		}
		return m.checkForUpdate()
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	v4Exists, err := util.IsFileExists(m.DataFilePathV4)
	if err != nil {
		return err
	}
	v6Exists, err := util.IsFileExists(m.DataFilePathV6)
	if err != nil {
		return err
	}
	if !v4Exists || !v6Exists {
		common.VerboseOutput("Cloudflare IP ranges file does not exist.")
		if m.UpdatePolicy.NoUpdate {
			return errors.New("Cloudflare IP ranges file does not exist and --no-update is enabled")
		}
		return m.downloadData()
	}

	return m.checkForUpdate()
}

func (m *IpDataManagerCloudflare) checkForUpdate() error {
	policy := m.UpdatePolicy
	if policy.NoUpdate {
		common.VerboseOutput("Cloudflare IP ranges update check skipped.")
		return nil
	}
	if metadataManager.IsUpdateCheckFresh(time.Now(), policy.EffectiveTTL()) {
		common.VerboseOutput("Cloudflare IP ranges update check skipped; cache is fresh.")
		return nil
	}

	data, err := m.fetchData()
	if err != nil {
		return m.useCacheIfUpdateFails(err)
	}
	signature, err := data.signature()
	if err != nil {
		return err
	}
	if metadataManager.IsSignatureExpired(signature) {
		common.VerboseOutput("Cloudflare IP ranges are outdated. Updating to the latest version...")
		return m.useCacheIfUpdateFails(m.writeData(data))
	}
	if err := metadataManager.MarkChecked(time.Now()); err != nil {
		return m.useCacheIfUpdateFails(util.ErrorWithInfo(err, "error writing metadata"))
	}
	common.VerboseOutput("Cloudflare IP ranges are up-to-date.")

	return nil
}

func (m *IpDataManagerCloudflare) LoadIpData() (*IpRangeDataCloudflare, error) {
	if !m.IpRange.IsEmpty() {
		return &m.IpRange, nil
	}

	data, err := m.readDataFile()
	if err != nil {
		return nil, err
	}
	m.IpRange = *data
	return &m.IpRange, nil
}

func readCIDRLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error opening data file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error reading data file")
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
	return cidrs, nil
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

func (m *IpDataManagerCloudflare) useCacheIfUpdateFails(updateErr error) error {
	if updateErr == nil {
		return nil
	}
	if _, err := m.LoadIpData(); err != nil {
		return errors.Join(updateErr, err)
	}
	util.PrintErrorTrace(util.ErrorWithInfo(updateErr, "using cached Cloudflare IP ranges"))
	return nil
}
