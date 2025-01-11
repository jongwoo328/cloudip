package aws

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

type IpDataManagerAws struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataAws
}

type IpRangeDataAws struct {
	SyncToken  string `json:"syncToken"`
	CreateDate string `json:"createDate"`
	Prefixes   []struct {
		IpPrefix           string `json:"ip_prefix"`
		Region             string `json:"region"`
		Service            string `json:"service"`
		NetworkBorderGroup string `json:"network_border_group"`
	} `json:"prefixes"`
	Ipv6Prefixes []struct {
		Ipv6Prefix         string `json:"ipv6_prefix"`
		Region             string `json:"region"`
		Service            string `json:"service"`
		NetworkBorderGroup string `json:"network_border_group"`
	} `json:"ipv6_prefixes"`
}

func (ipRange IpRangeDataAws) IsEmpty() bool {
	return ipRange.SyncToken == "" &&
		ipRange.CreateDate == "" &&
		len(ipRange.Prefixes) == 0 &&
		len(ipRange.Ipv6Prefixes) == 0
}

func (ipDataManagerAws *IpDataManagerAws) GetLastModifiedUpstream() (time.Time, error) {
	resp, err := http.Head(ipDataManagerAws.DataURI)
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

func (ipDataManagerAws *IpDataManagerAws) DownloadData() error {
	common.VerboseOutput("Downloading AWS IP ranges...")
	if ipDataManagerAws.DataURI == "" {
		return errors.New("cannot get DataURI")
	}
	resp, err := http.Get(ipDataManagerAws.DataURI)
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

	dataFile, err := os.OpenFile(ipDataManagerAws.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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
		common.VerboseOutput(fmt.Sprintf("AWS IP ranges updated [%s]", util.FormatToTimestamp(currentLastModified)))
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

func (ipDataManagerAws *IpDataManagerAws) EnsureDataFile() error {
	if !util.IsFileExists(DataFilePathAws) {
		common.VerboseOutput("AWS IP ranges file not exists.")
		// Download the AWS IP ranges file
		err := ipDataManagerAws.DownloadData()
		return err
	}
	if isExpired() {
		common.VerboseOutput("AWS IP ranges are outdated. Updating to the latest version...")
		// update the file
		err := ipDataManagerAws.DownloadData()
		return err
	}
	common.VerboseOutput("AWS IP ranges are up-to-date.")

	return nil
}

func (ipDataManagerAws *IpDataManagerAws) LoadIpData() *IpRangeDataAws {
	if !ipDataManagerAws.IpRange.IsEmpty() {
		return &ipDataManagerAws.IpRange
	}

	awsIpRangeData := IpRangeDataAws{}
	ipDataFile, err := os.Open(ipDataManagerAws.DataFilePath)
	if err != nil {
		err = util.ErrorWithInfo(err, "Error opening data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}
	err = util.HandleJSON(ipDataFile, &awsIpRangeData, "read")
	if err != nil {
		err = util.ErrorWithInfo(err, "Error reading data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}

	ipDataManagerAws.IpRange = awsIpRangeData
	return &ipDataManagerAws.IpRange
}

var ipDataManagerAws = &IpDataManagerAws{
	DataURI:      getDaraUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathAws,
	IpRange:      IpRangeDataAws{},
}
