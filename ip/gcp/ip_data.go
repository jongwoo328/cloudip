package gcp

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	resp, err := http.Head(ipDataManagerGcp.DataURI)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error checking metadata file expiration"))
		return time.Time{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error closing response body"))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err := util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error checking metadata file expiration")
		util.PrintErrorTrace(err)
		return time.Time{}, err
	}

	lastModified := resp.Header.Get("Last-Modified")
	lastModifiedDate, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		err := util.ErrorWithInfo(err, "Error parsing Date header")
		util.PrintErrorTrace(err)
		return time.Time{}, err
	}

	return lastModifiedDate, nil
}

func (ipDataManagerGcp *IpDataManagerGcp) DownloadData() error {
	common.VerboseOutput("Downloading GCP IP ranges...")
	if ipDataManagerGcp.DataURI == "" {
		return errors.New("cannot get DataURI")
	}
	resp, err := http.Get(ipDataManagerGcp.DataURI)
	if err != nil {
		err = util.ErrorWithInfo(err, "Error downloading data file")
		util.PrintErrorTrace(err)
		return err
	}

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		err = util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error downloading data file")
		util.PrintErrorTrace(err)
		return err
	}

	dataFile, err := os.OpenFile(ipDataManagerGcp.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		err = util.ErrorWithInfo(err, "Error opening data file")
		util.PrintErrorTrace(err)
		return err
	}

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		err = util.ErrorWithInfo(err, "Error saving data file")
		util.PrintErrorTrace(err)
		return err
	}

	currentLastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		err = util.ErrorWithInfo(err, "Error parsing Date header")
		util.PrintErrorTrace(err)
		return err
	}

	if metadataManager.Metadata.LastModified != currentLastModified.Unix() {
		metadata := common.CloudMetadata{
			Type:         common.AWS,
			LastModified: currentLastModified.Unix(),
		}
		if err := writeMetadata(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "Error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput(fmt.Sprintf("GCP IP ranges updated [%s]", util.FormatToTimestamp(currentLastModified)))
	}

	defer func() {
		if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(networkCloseErr, "Error closing response body"))
		}
		if fileCloseErr := dataFile.Close(); fileCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(fileCloseErr, "Error closing data file"))
		}
	}()

	return nil
}

func (ipDataManagerGcp *IpDataManagerGcp) EnsureDataFile() error {
	if !util.IsFileExists(DataFilePathGcp) {
		common.VerboseOutput("GCP IP ranges file not exists.")
		err := ipDataManagerGcp.DownloadData()
		return err
	}
	if isExpired() {
		common.VerboseOutput("GCP IP ranges are outdated. Updating to the latest version...")
		err := ipDataManagerGcp.DownloadData()
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
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error opening data file"))
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}
	err = util.HandleJSON(ipDataFile, &gcpIpRangeData, "read")
	if err != nil {
		err = util.ErrorWithInfo(err, "Error reading data file")
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
