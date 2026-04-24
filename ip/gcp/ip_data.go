package gcp

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"log"
	"os"
)

type IpDataManagerGcp struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataGcp
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
	return &gcpIpRangeData, nil
}

func (ipDataManagerGcp *IpDataManagerGcp) writeData(gcpIpRangeData *IpRangeDataGcp) error {
	if gcpIpRangeData.SyncToken == "" {
		return errors.New("cannot get syncToken")
	}

	ipDataFile, err := os.OpenFile(ipDataManagerGcp.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		err = util.ErrorWithInfo(err, "error opening data file")
		util.PrintErrorTrace(err)
		return err
	}
	defer ipDataFile.Close()

	if err := util.WriteJSON(ipDataFile, gcpIpRangeData); err != nil {
		err = util.ErrorWithInfo(err, "error writing data file")
		util.PrintErrorTrace(err)
		return err
	}

	if metadataManager.IsSignatureExpired(gcpIpRangeData.SyncToken) {
		metadata := common.CloudMetadata{
			Type:      common.GCP,
			Signature: gcpIpRangeData.SyncToken,
		}
		if err := metadataManager.Write(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput(fmt.Sprintf("GCP IP ranges updated [%s]", gcpIpRangeData.CreationTime))
	}

	return nil
}

func (ipDataManagerGcp *IpDataManagerGcp) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	if !util.IsFileExists(ipDataManagerGcp.DataFilePath) {
		common.VerboseOutput("GCP IP ranges file not exists.")
		err := ipDataManagerGcp.downloadData()
		return err
	}

	gcpIpRangeData, err := ipDataManagerGcp.fetchData()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error getting signature from GCP server"))
		return nil
	}
	if gcpIpRangeData.SyncToken == "" {
		return errors.New("cannot get syncToken")
	}
	if metadataManager.IsSignatureExpired(gcpIpRangeData.SyncToken) {
		common.VerboseOutput("GCP IP ranges are outdated. Updating to the latest version...")
		return ipDataManagerGcp.writeData(gcpIpRangeData)
	}
	common.VerboseOutput("GCP IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerGcp *IpDataManagerGcp) LoadIpData() *IpRangeDataGcp {
	if !ipDataManagerGcp.IpRange.IsEmpty() {
		return &ipDataManagerGcp.IpRange
	}

	gcpIpRangeData, err := ipDataManagerGcp.readDataFile()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error opening data file"))
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}

	ipDataManagerGcp.IpRange = *gcpIpRangeData
	return &ipDataManagerGcp.IpRange
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
