package azure

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

type IpDataManagerAzure struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataAzure
}

type IpRangeDataAzure struct {
	ChangeNumber int    `json:"changeNumber"`
	Cloud        string `json:"cloud"`
	Values       []struct {
		Name       string `json:"name"`
		Id         string `json:"id"`
		Properties struct {
			ChangeNumber    int      `json:"changeNumber"`
			Region          string   `json:"region"`
			RegionId        int      `json:"regionId"`
			Platform        string   `json:"platform"`
			SystemService   string   `json:"systemService"`
			AddressPrefixes []string `json:"addressPrefixes"`
			NetworkFeatures []string `json:"networkFeatures"`
		} `json:"properties"`
	} `json:"values"`
}

func (ipRange IpRangeDataAzure) IsEmpty() bool {
	return ipRange.ChangeNumber == 0 &&
		len(ipRange.Values) == 0
}

func (ipDataManagerAzure *IpDataManagerAzure) GetLastModifiedUpstream() (time.Time, error) {
	resp, err := http.Head(ipDataManagerAzure.DataURI)
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

func (ipDataManagerAzure *IpDataManagerAzure) DownloadData() error {
	common.VerboseOutput("Downloading Azure IP ranges...")
	if ipDataManagerAzure.DataURI == "" {
		return errors.New("cannot get DataURI")
	}
	resp, err := http.Get(ipDataManagerAzure.DataURI)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error downloading data file"))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error downloading data file")
		util.PrintErrorTrace(err)
		return err
	}
	dataFile, err := os.OpenFile(ipDataManagerAzure.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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
			Type:         common.Azure,
			LastModified: currentLastModified.Unix(),
		}
		if err := writeMetadata(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "Error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput(fmt.Sprintf("Azure IP ranges updated [%s]", util.FormatToTimestamp(currentLastModified)))

	}

	defer func() {
		if networkCloseEre := resp.Body.Close(); networkCloseEre != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(networkCloseEre, "Error closing response body"))
		}
		if fileCloseErr := dataFile.Close(); fileCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(fileCloseErr, "Error closing data file"))
		}
	}()

	return nil
}

func (ipDataManagerAzure *IpDataManagerAzure) EnsureDataFile() error {
	if !util.IsFileExists(DataFilePathAzure) {
		common.VerboseOutput("Azure IP ranged file not exists.")
		err := ipDataManagerAzure.DownloadData()
		return err
	}
	if isExpired() {
		common.VerboseOutput("Azure IP ranged are outdated. Updating to the latest version...")
		err := ipDataManagerAzure.DownloadData()
		return err
	}
	common.VerboseOutput("Azure IP ranged are up-to-date.")

	return nil
}

func (ipDataManagerAzure *IpDataManagerAzure) LoadIpData() *IpRangeDataAzure {
	if !ipDataManagerAzure.IpRange.IsEmpty() {
		return &ipDataManagerAzure.IpRange
	}

	azureIpRangeData := IpRangeDataAzure{}
	ipDataFile, err := os.Open(ipDataManagerAzure.DataFilePath)
	if err != nil {
		err = util.ErrorWithInfo(err, "Error loading data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}
	err = util.HandleJSON(ipDataFile, &azureIpRangeData, "read")
	if err != nil {
		err = util.ErrorWithInfo(err, "Error reading data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}

	ipDataManagerAzure.IpRange = azureIpRangeData
	return &ipDataManagerAzure.IpRange
}

var ipDataManagerAzure = &IpDataManagerAzure{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathAzure,
	IpRange:      IpRangeDataAzure{},
}
