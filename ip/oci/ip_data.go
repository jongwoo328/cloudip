package oci

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"os"
	"time"
)

type IpDataManagerOCI struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataOCI
	UpdatePolicy common.UpdatePolicy
}

type CIDR struct {
	CIDR string   `json:"cidr"`
	Tags []string `json:"tags,omitempty"`
}

type IpRangeDataOCI struct {
	LastUpdatedTimestamp string `json:"last_updated_timestamp"`
	Regions              []struct {
		Region    string `json:"region"`
		CIDRs     []CIDR `json:"cidrs"`
		IPv6CIDRs []CIDR `json:"ipv6_cidrs,omitempty"`
	} `json:"regions"`
}

func (data IpRangeDataOCI) IsEmpty() bool {
	return data.LastUpdatedTimestamp == "" && len(data.Regions) == 0
}

func (ipDataManagerOCI *IpDataManagerOCI) downloadData() error {
	common.VerboseOutput("Downloading OCI IP ranges...")
	if ipDataManagerOCI.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	ociIpRangeData, err := ipDataManagerOCI.fetchData()
	if err != nil {
		return err
	}
	return ipDataManagerOCI.writeData(ociIpRangeData)
}

func (ipDataManagerOCI *IpDataManagerOCI) fetchData() (*IpRangeDataOCI, error) {
	ociIpRangeData := IpRangeDataOCI{}
	_, err := util.DownloadJSONFromUrl(ipDataManagerOCI.DataURI, &ociIpRangeData)
	if err != nil {
		return nil, err
	}
	if err := ociIpRangeData.validate(); err != nil {
		return nil, err
	}
	return &ociIpRangeData, nil
}

func (ipDataManagerOCI *IpDataManagerOCI) writeData(ociIpRangeData *IpRangeDataOCI) error {
	if err := ociIpRangeData.validate(); err != nil {
		return err
	}
	if err := util.WriteJSONFile(ipDataManagerOCI.DataFilePath, ociIpRangeData); err != nil {
		return err
	}

	signatureExpired := metadataManager.IsSignatureExpired(ociIpRangeData.LastUpdatedTimestamp)
	metadata := common.CloudMetadata{
		Type:        common.OCI,
		Signature:   ociIpRangeData.LastUpdatedTimestamp,
		LastChecked: time.Now().Unix(),
	}
	if err := metadataManager.Write(&metadata); err != nil {
		err = util.ErrorWithInfo(err, "error writing metadata")
		util.PrintErrorTrace(err)
		return err
	}
	ipDataManagerOCI.IpRange = *ociIpRangeData
	if signatureExpired {
		common.VerboseOutput(fmt.Sprintf("OCI IP ranges updated [%s]", ociIpRangeData.LastUpdatedTimestamp))
	}

	return nil
}

func (ipDataManagerOCI *IpDataManagerOCI) SetUpdatePolicy(policy common.UpdatePolicy) {
	ipDataManagerOCI.UpdatePolicy = policy
}

func (ipDataManagerOCI *IpDataManagerOCI) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	exists, err := util.IsFileExists(ipDataManagerOCI.DataFilePath)
	if err != nil {
		return err
	}
	if !exists {
		common.VerboseOutput("OCI IP ranges file does not exist.")
		if ipDataManagerOCI.UpdatePolicy.NoUpdate {
			return errors.New("OCI IP ranges file does not exist and --no-update is enabled")
		}
		err := ipDataManagerOCI.downloadData()
		return err
	}

	policy := ipDataManagerOCI.UpdatePolicy
	if policy.NoUpdate {
		common.VerboseOutput("OCI IP ranges update check skipped.")
		return nil
	}
	if metadataManager.IsUpdateCheckFresh(time.Now(), policy.EffectiveTTL()) {
		common.VerboseOutput("OCI IP ranges update check skipped; cache is fresh.")
		return nil
	}

	ociIpRangeData, err := ipDataManagerOCI.fetchData()
	if err != nil {
		return ipDataManagerOCI.useCacheIfUpdateFails(err)
	}
	if metadataManager.IsSignatureExpired(ociIpRangeData.LastUpdatedTimestamp) {
		common.VerboseOutput("OCI IP ranges are outdated. Updating to the latest version...")
		return ipDataManagerOCI.useCacheIfUpdateFails(ipDataManagerOCI.writeData(ociIpRangeData))
	}
	if err := metadataManager.MarkChecked(time.Now()); err != nil {
		return ipDataManagerOCI.useCacheIfUpdateFails(util.ErrorWithInfo(err, "error writing metadata"))
	}
	common.VerboseOutput("OCI IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerOCI *IpDataManagerOCI) LoadIpData() (*IpRangeDataOCI, error) {
	if !ipDataManagerOCI.IpRange.IsEmpty() {
		return &ipDataManagerOCI.IpRange, nil
	}

	ociIpRangeData, err := ipDataManagerOCI.readDataFile()
	if err != nil {
		return nil, err
	}

	if err := ociIpRangeData.validate(); err != nil {
		return nil, err
	}
	ipDataManagerOCI.IpRange = *ociIpRangeData
	return &ipDataManagerOCI.IpRange, nil
}

func (ipDataManagerOCI *IpDataManagerOCI) readDataFile() (*IpRangeDataOCI, error) {
	ociIpRangeData := IpRangeDataOCI{}
	ipDataFile, err := os.Open(ipDataManagerOCI.DataFilePath)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error opening data file")
	}
	defer ipDataFile.Close()

	err = util.ReadJSON(ipDataFile, &ociIpRangeData)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error reading data file")
	}

	return &ociIpRangeData, nil
}

var ipDataManagerOCI = &IpDataManagerOCI{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathOCI,
	IpRange:      IpRangeDataOCI{},
}

func (m *IpDataManagerOCI) useCacheIfUpdateFails(updateErr error) error {
	if updateErr == nil {
		return nil
	}
	if _, err := m.LoadIpData(); err != nil {
		return errors.Join(updateErr, err)
	}
	util.PrintErrorTrace(util.ErrorWithInfo(updateErr, "using cached OCI IP ranges"))
	return nil
}
