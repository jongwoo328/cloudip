package digitalocean

import (
	"bytes"
	"cloudip/common"
	"cloudip/util"
	"errors"
	"os"
	"time"
)

type IpDataManagerDigitalOcean struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataDigitalOcean
	UpdatePolicy common.UpdatePolicy
}

type IpRangeDataDigitalOcean struct {
	CIDRs []string `json:"cidrs"`
}

func (data IpRangeDataDigitalOcean) IsEmpty() bool { return len(data.CIDRs) == 0 }

func (ipDataManagerDigitalOcean *IpDataManagerDigitalOcean) downloadData() error {
	common.VerboseOutput("Downloading DigitalOcean IP ranges...")
	if ipDataManagerDigitalOcean.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	digitaloceanIpRangeData, err := ipDataManagerDigitalOcean.fetchData()
	if err != nil {
		return err
	}
	return ipDataManagerDigitalOcean.writeData(digitaloceanIpRangeData)
}

func (m *IpDataManagerDigitalOcean) fetchData() (*IpRangeDataDigitalOcean, error) {
	body, err := util.DownloadFromUrl(m.DataURI)
	if err != nil {
		return nil, err
	}
	return decodeGeofeed(bytes.NewReader(body))
}

func (ipDataManagerDigitalOcean *IpDataManagerDigitalOcean) writeData(digitaloceanIpRangeData *IpRangeDataDigitalOcean) error {
	if err := digitaloceanIpRangeData.validate(); err != nil {
		return err
	}
	if err := util.WriteJSONFile(ipDataManagerDigitalOcean.DataFilePath, digitaloceanIpRangeData); err != nil {
		return err
	}

	signatureExpired := metadataManager.IsSignatureExpired(digitaloceanIpRangeData.signature())
	metadata := common.CloudMetadata{
		Type:        common.DigitalOcean,
		Signature:   digitaloceanIpRangeData.signature(),
		LastChecked: time.Now().Unix(),
	}
	if err := metadataManager.Write(&metadata); err != nil {
		err = util.ErrorWithInfo(err, "error writing metadata")
		util.PrintErrorTrace(err)
		return err
	}
	ipDataManagerDigitalOcean.IpRange = *digitaloceanIpRangeData
	if signatureExpired {
		common.VerboseOutput("DigitalOcean IP ranges updated")
	}

	return nil
}

func (ipDataManagerDigitalOcean *IpDataManagerDigitalOcean) SetUpdatePolicy(policy common.UpdatePolicy) {
	ipDataManagerDigitalOcean.UpdatePolicy = policy
}

func (ipDataManagerDigitalOcean *IpDataManagerDigitalOcean) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	exists, err := util.IsFileExists(ipDataManagerDigitalOcean.DataFilePath)
	if err != nil {
		return err
	}
	if !exists {
		common.VerboseOutput("DigitalOcean IP ranges file does not exist.")
		if ipDataManagerDigitalOcean.UpdatePolicy.NoUpdate {
			return errors.New("DigitalOcean IP ranges file does not exist and --no-update is enabled")
		}
		err := ipDataManagerDigitalOcean.downloadData()
		return err
	}

	policy := ipDataManagerDigitalOcean.UpdatePolicy
	if policy.NoUpdate {
		common.VerboseOutput("DigitalOcean IP ranges update check skipped.")
		return nil
	}
	if metadataManager.IsUpdateCheckFresh(time.Now(), policy.EffectiveTTL()) {
		common.VerboseOutput("DigitalOcean IP ranges update check skipped; cache is fresh.")
		return nil
	}

	digitaloceanIpRangeData, err := ipDataManagerDigitalOcean.fetchData()
	if err != nil {
		return ipDataManagerDigitalOcean.useCacheIfUpdateFails(err)
	}
	if metadataManager.IsSignatureExpired(digitaloceanIpRangeData.signature()) {
		common.VerboseOutput("DigitalOcean IP ranges are outdated. Updating to the latest version...")
		return ipDataManagerDigitalOcean.useCacheIfUpdateFails(ipDataManagerDigitalOcean.writeData(digitaloceanIpRangeData))
	}
	if err := metadataManager.MarkChecked(time.Now()); err != nil {
		return ipDataManagerDigitalOcean.useCacheIfUpdateFails(util.ErrorWithInfo(err, "error writing metadata"))
	}
	common.VerboseOutput("DigitalOcean IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerDigitalOcean *IpDataManagerDigitalOcean) LoadIpData() (*IpRangeDataDigitalOcean, error) {
	if !ipDataManagerDigitalOcean.IpRange.IsEmpty() {
		return &ipDataManagerDigitalOcean.IpRange, nil
	}

	digitaloceanIpRangeData, err := ipDataManagerDigitalOcean.readDataFile()
	if err != nil {
		return nil, err
	}

	if err := digitaloceanIpRangeData.validate(); err != nil {
		return nil, err
	}
	ipDataManagerDigitalOcean.IpRange = *digitaloceanIpRangeData
	return &ipDataManagerDigitalOcean.IpRange, nil
}

func (ipDataManagerDigitalOcean *IpDataManagerDigitalOcean) readDataFile() (*IpRangeDataDigitalOcean, error) {
	digitaloceanIpRangeData := IpRangeDataDigitalOcean{}
	ipDataFile, err := os.Open(ipDataManagerDigitalOcean.DataFilePath)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error opening data file")
	}
	defer ipDataFile.Close()

	err = util.ReadJSON(ipDataFile, &digitaloceanIpRangeData)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error reading data file")
	}

	return &digitaloceanIpRangeData, nil
}

var ipDataManagerDigitalOcean = &IpDataManagerDigitalOcean{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathDigitalOcean,
	IpRange:      IpRangeDataDigitalOcean{},
}

func (m *IpDataManagerDigitalOcean) useCacheIfUpdateFails(updateErr error) error {
	if updateErr == nil {
		return nil
	}
	if _, err := m.LoadIpData(); err != nil {
		return errors.Join(updateErr, err)
	}
	util.PrintErrorTrace(util.ErrorWithInfo(updateErr, "using cached DigitalOcean IP ranges"))
	return nil
}
