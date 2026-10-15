package gcp

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"os"
	"time"
)

type IpDataManagerGcp struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataGcp
	UpdatePolicy common.UpdatePolicy
}

type IpRangeDataGcp struct {
	SyncToken    string `json:"syncToken"`
	CreationTime string `json:"creationTime"`
	Prefixes     []struct {
		Ipv4Prefix string `json:"ipv4Prefix"`
		Ipv6Prefix string `json:"ipv6Prefix"`
		Service    string `json:"service"`
		Scope      string `json:"scope"`
	} `json:"prefixes"`
}

func (ipRange IpRangeDataGcp) IsEmpty() bool {
	return ipRange.SyncToken == "" &&
		ipRange.CreationTime == "" &&
		len(ipRange.Prefixes) == 0
}

func (ipDataManagerGcp *IpDataManagerGcp) downloadData() error {
	common.VerboseOutput("Downloading GCP IP ranges...")
	if ipDataManagerGcp.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	gcpIpRangeData, err := ipDataManagerGcp.fetchData()
	if err != nil {
		return err
	}
	return ipDataManagerGcp.writeData(gcpIpRangeData)
}

func (ipDataManagerGcp *IpDataManagerGcp) fetchData() (*IpRangeDataGcp, error) {
	gcpIpRangeData := IpRangeDataGcp{}
	_, err := util.DownloadJSONFromUrl(ipDataManagerGcp.DataURI, &gcpIpRangeData)
	if err != nil {
		return nil, err
	}
	if err := gcpIpRangeData.validate(); err != nil {
		return nil, err
	}
	return &gcpIpRangeData, nil
}

func (ipDataManagerGcp *IpDataManagerGcp) writeData(gcpIpRangeData *IpRangeDataGcp) error {
	if err := gcpIpRangeData.validate(); err != nil {
		return err
	}
	if err := util.WriteJSONFile(ipDataManagerGcp.DataFilePath, gcpIpRangeData); err != nil {
		return err
	}

	signatureExpired := metadataManager.IsSignatureExpired(gcpIpRangeData.SyncToken)
	metadata := common.CloudMetadata{
		Type:        common.GCP,
		Signature:   gcpIpRangeData.SyncToken,
		LastChecked: time.Now().Unix(),
	}
	if err := metadataManager.Write(&metadata); err != nil {
		err = util.ErrorWithInfo(err, "error writing metadata")
		util.PrintErrorTrace(err)
		return err
	}
	if signatureExpired {
		common.VerboseOutput(fmt.Sprintf("GCP IP ranges updated [%s]", gcpIpRangeData.CreationTime))
	}

	return nil
}

func (ipDataManagerGcp *IpDataManagerGcp) SetUpdatePolicy(policy common.UpdatePolicy) {
	ipDataManagerGcp.UpdatePolicy = policy
}

func (ipDataManagerGcp *IpDataManagerGcp) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	exists, err := util.IsFileExists(ipDataManagerGcp.DataFilePath)
	if err != nil {
		return err
	}
	if !exists {
		common.VerboseOutput("GCP IP ranges file does not exist.")
		if ipDataManagerGcp.UpdatePolicy.NoUpdate {
			return errors.New("GCP IP ranges file does not exist and --no-update is enabled")
		}
		err := ipDataManagerGcp.downloadData()
		return err
	}

	policy := ipDataManagerGcp.UpdatePolicy
	if policy.NoUpdate {
		common.VerboseOutput("GCP IP ranges update check skipped.")
		return nil
	}
	if metadataManager.IsUpdateCheckFresh(time.Now(), policy.EffectiveTTL()) {
		common.VerboseOutput("GCP IP ranges update check skipped; cache is fresh.")
		return nil
	}

	gcpIpRangeData, err := ipDataManagerGcp.fetchData()
	if err != nil {
		return ipDataManagerGcp.useCacheIfUpdateFails(err)
	}
	if metadataManager.IsSignatureExpired(gcpIpRangeData.SyncToken) {
		common.VerboseOutput("GCP IP ranges are outdated. Updating to the latest version...")
		return ipDataManagerGcp.useCacheIfUpdateFails(ipDataManagerGcp.writeData(gcpIpRangeData))
	}
	if err := metadataManager.MarkChecked(time.Now()); err != nil {
		return ipDataManagerGcp.useCacheIfUpdateFails(util.ErrorWithInfo(err, "error writing metadata"))
	}
	common.VerboseOutput("GCP IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerGcp *IpDataManagerGcp) LoadIpData() (*IpRangeDataGcp, error) {
	if !ipDataManagerGcp.IpRange.IsEmpty() {
		return &ipDataManagerGcp.IpRange, nil
	}

	gcpIpRangeData, err := ipDataManagerGcp.readDataFile()
	if err != nil {
		return nil, err
	}

	if err := gcpIpRangeData.validate(); err != nil {
		return nil, err
	}
	ipDataManagerGcp.IpRange = *gcpIpRangeData
	return &ipDataManagerGcp.IpRange, nil
}

func (ipDataManagerGcp *IpDataManagerGcp) readDataFile() (*IpRangeDataGcp, error) {
	gcpIpRangeData := IpRangeDataGcp{}
	ipDataFile, err := os.Open(ipDataManagerGcp.DataFilePath)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error opening data file")
	}
	defer ipDataFile.Close()

	err = util.ReadJSON(ipDataFile, &gcpIpRangeData)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error reading data file")
	}

	return &gcpIpRangeData, nil
}

var ipDataManagerGcp = &IpDataManagerGcp{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathGcp,
	IpRange:      IpRangeDataGcp{},
}

func (m *IpDataManagerGcp) useCacheIfUpdateFails(updateErr error) error {
	if updateErr == nil {
		return nil
	}
	if _, err := m.LoadIpData(); err != nil {
		return errors.Join(updateErr, err)
	}
	util.PrintErrorTrace(util.ErrorWithInfo(updateErr, "using cached GCP IP ranges"))
	return nil
}
