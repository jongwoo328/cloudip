package gcp

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
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

func (ipDataManagerGcp *IpDataManagerGcp) GetLastModifiedUpstream() (time.Time, error) {
	headers, err := util.GetHeadRequestHeader(ipDataManagerGcp.DataURI)
	if err != nil {
		err = util.ErrorWithInfo(err, "error getting header from request")
		util.PrintErrorTrace(err)
		return time.Time{}, err
	}

	lastModified := headers.Get("Last-Modified")
	lastModifiedDate, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		err := util.ErrorWithInfo(err, "error parsing Date header")
		util.PrintErrorTrace(err)
		return time.Time{}, err
	}

	return lastModifiedDate, nil
}

func (ipDataManagerGcp *IpDataManagerGcp) downloadData() error {
	common.VerboseOutput("Downloading GCP IP ranges...")
	if ipDataManagerGcp.DataURI == "" {
		return errors.New("cannot get DataURI")
	}
	err := util.DownloadFromUrlToFile(ipDataManagerGcp.DataURI, ipDataManagerGcp.DataFilePath)
	if err != nil {
		return err
	}

	headers, err := util.GetHeadRequestHeader(ipDataManagerGcp.DataURI)
	if err != nil {
		err = util.ErrorWithInfo(err, "error getting header from request")
		util.PrintErrorTrace(err)
		return err
	}
	currentLastModified, err := time.Parse(time.RFC1123, headers.Get("Last-Modified"))
	if err != nil {
		err = util.ErrorWithInfo(err, "error parsing Date header")
		util.PrintErrorTrace(err)
		return err
	}

	if metadataManager.Metadata.LastModified != currentLastModified.Unix() {
		metadata := common.CloudMetadata{
			Type:         common.GCP,
			LastModified: currentLastModified.Unix(),
		}
		if err := metadataManager.Write(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput(fmt.Sprintf("GCP IP ranges updated [%s]", util.FormatToTimestamp(currentLastModified)))
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

	if !util.IsFileExists(DataFilePathGcp) {
		common.VerboseOutput("GCP IP ranges file not exists.")
		err := ipDataManagerGcp.downloadData()
		return err
	}
	if isExpired() {
		common.VerboseOutput("GCP IP ranges are outdated. Updating to the latest version...")
		err := ipDataManagerGcp.downloadData()
		return err
	}
	common.VerboseOutput("GCP IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerGcp *IpDataManagerGcp) LoadIpData() *IpRangeDataGcp {
	if !ipDataManagerGcp.IpRange.IsEmpty() {
		return &ipDataManagerGcp.IpRange
	}

	gcpIpRangeData := IpRangeDataGcp{}
	ipDataFile, err := os.Open(ipDataManagerGcp.DataFilePath)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "error opening data file"))
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}
	err = util.ReadJSON(ipDataFile, &gcpIpRangeData)
	if err != nil {
		err = util.ErrorWithInfo(err, "error reading data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}

	ipDataManagerGcp.IpRange = gcpIpRangeData
	return &ipDataManagerGcp.IpRange
}

var ipDataManagerGcp = &IpDataManagerGcp{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathGcp,
	IpRange:      IpRangeDataGcp{},
}
