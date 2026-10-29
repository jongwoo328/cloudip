package vultr

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"os"
	"time"
)

type IpDataManagerVultr struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataVultr
	UpdatePolicy common.UpdatePolicy
}

type IpRangeDataVultr struct {
	Updated string `json:"updated"`
	Subnets []struct {
		IPPrefix string `json:"ip_prefix"`
	} `json:"subnets"`
}

func (data IpRangeDataVultr) IsEmpty() bool {
	return data.Updated == "" && len(data.Subnets) == 0
}

func (ipDataManagerVultr *IpDataManagerVultr) downloadData() error {
	common.VerboseOutput("Downloading Vultr IP ranges...")
	if ipDataManagerVultr.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	vultrIpRangeData, err := ipDataManagerVultr.fetchData()
	if err != nil {
		return err
	}
	return ipDataManagerVultr.writeData(vultrIpRangeData)
}

func (ipDataManagerVultr *IpDataManagerVultr) fetchData() (*IpRangeDataVultr, error) {
	vultrIpRangeData := IpRangeDataVultr{}
	_, err := util.DownloadJSONFromUrl(ipDataManagerVultr.DataURI, &vultrIpRangeData)
	if err != nil {
		return nil, err
	}
	if err := vultrIpRangeData.validate(); err != nil {
		return nil, err
	}
	return &vultrIpRangeData, nil
}

func (ipDataManagerVultr *IpDataManagerVultr) writeData(vultrIpRangeData *IpRangeDataVultr) error {
	if err := vultrIpRangeData.validate(); err != nil {
		return err
	}
	if err := util.WriteJSONFile(ipDataManagerVultr.DataFilePath, vultrIpRangeData); err != nil {
		return err
	}

	signatureExpired := metadataManager.IsSignatureExpired(vultrIpRangeData.Updated)
	metadata := common.CloudMetadata{
		Type:        common.Vultr,
		Signature:   vultrIpRangeData.Updated,
		LastChecked: time.Now().Unix(),
	}
	if err := metadataManager.Write(&metadata); err != nil {
		err = util.ErrorWithInfo(err, "error writing metadata")
		util.PrintErrorTrace(err)
		return err
	}
	ipDataManagerVultr.IpRange = *vultrIpRangeData
	if signatureExpired {
		common.VerboseOutput(fmt.Sprintf("Vultr IP ranges updated [%s]", vultrIpRangeData.Updated))
	}

	return nil
}

func (ipDataManagerVultr *IpDataManagerVultr) SetUpdatePolicy(policy common.UpdatePolicy) {
	ipDataManagerVultr.UpdatePolicy = policy
}

func (ipDataManagerVultr *IpDataManagerVultr) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	exists, err := util.IsFileExists(ipDataManagerVultr.DataFilePath)
	if err != nil {
		return err
	}
	if !exists {
		common.VerboseOutput("Vultr IP ranges file does not exist.")
		if ipDataManagerVultr.UpdatePolicy.NoUpdate {
			return errors.New("Vultr IP ranges file does not exist and --no-update is enabled")
		}
		err := ipDataManagerVultr.downloadData()
		return err
	}

	policy := ipDataManagerVultr.UpdatePolicy
	if policy.NoUpdate {
		common.VerboseOutput("Vultr IP ranges update check skipped.")
		return nil
	}
	if metadataManager.IsUpdateCheckFresh(time.Now(), policy.EffectiveTTL()) {
		common.VerboseOutput("Vultr IP ranges update check skipped; cache is fresh.")
		return nil
	}

	vultrIpRangeData, err := ipDataManagerVultr.fetchData()
	if err != nil {
		return ipDataManagerVultr.useCacheIfUpdateFails(err)
	}
	if metadataManager.IsSignatureExpired(vultrIpRangeData.Updated) {
		common.VerboseOutput("Vultr IP ranges are outdated. Updating to the latest version...")
		return ipDataManagerVultr.useCacheIfUpdateFails(ipDataManagerVultr.writeData(vultrIpRangeData))
	}
	if err := metadataManager.MarkChecked(time.Now()); err != nil {
		return ipDataManagerVultr.useCacheIfUpdateFails(util.ErrorWithInfo(err, "error writing metadata"))
	}
	common.VerboseOutput("Vultr IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerVultr *IpDataManagerVultr) LoadIpData() (*IpRangeDataVultr, error) {
	if !ipDataManagerVultr.IpRange.IsEmpty() {
		return &ipDataManagerVultr.IpRange, nil
	}

	vultrIpRangeData, err := ipDataManagerVultr.readDataFile()
	if err != nil {
		return nil, err
	}

	if err := vultrIpRangeData.validate(); err != nil {
		return nil, err
	}
	ipDataManagerVultr.IpRange = *vultrIpRangeData
	return &ipDataManagerVultr.IpRange, nil
}

func (ipDataManagerVultr *IpDataManagerVultr) readDataFile() (*IpRangeDataVultr, error) {
	vultrIpRangeData := IpRangeDataVultr{}
	ipDataFile, err := os.Open(ipDataManagerVultr.DataFilePath)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error opening data file")
	}
	defer ipDataFile.Close()

	err = util.ReadJSON(ipDataFile, &vultrIpRangeData)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error reading data file")
	}

	return &vultrIpRangeData, nil
}

var ipDataManagerVultr = &IpDataManagerVultr{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathVultr,
	IpRange:      IpRangeDataVultr{},
}

func (m *IpDataManagerVultr) useCacheIfUpdateFails(updateErr error) error {
	if updateErr == nil {
		return nil
	}
	if _, err := m.LoadIpData(); err != nil {
		return errors.Join(updateErr, err)
	}
	util.PrintErrorTrace(util.ErrorWithInfo(updateErr, "using cached Vultr IP ranges"))
	return nil
}
