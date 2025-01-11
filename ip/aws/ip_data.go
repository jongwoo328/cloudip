package aws

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
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

func (ipDataManagerAws *IpDataManagerAws) downloadData() error {
	common.VerboseOutput("Downloading AWS IP ranges...")
	if ipDataManagerAws.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	err := util.DownloadFromUrlToFile(ipDataManagerAws.DataURI, ipDataManagerAws.DataFilePath)
	if err != nil {
		return err
	}

	headers, err := util.GetHeadRequestHeader(ipDataManagerAws.DataURI)
	if err != nil {
		err = util.ErrorWithInfo(err, "Error getting header from request")
		util.PrintErrorTrace(err)
		return err
	}
	currentLastModified, err := time.Parse(time.RFC1123, headers.Get("Last-Modified"))
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

	return nil
}

func (ipDataManagerAws *IpDataManagerAws) EnsureDataFile() error {
	if !util.IsFileExists(DataFilePathAws) {
		common.VerboseOutput("AWS IP ranges file not exists.")
		// Download the AWS IP ranges file
		err := ipDataManagerAws.downloadData()
		return err
	}
	if isExpired() {
		common.VerboseOutput("AWS IP ranges are outdated. Updating to the latest version...")
		// update the file
		err := ipDataManagerAws.downloadData()
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
