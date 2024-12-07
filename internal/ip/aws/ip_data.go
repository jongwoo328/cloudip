package aws

import (
	"cloudip/internal"
	"cloudip/internal/util"
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
	IpRange      IpRangeData
}

type IpRangeData struct {
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

func (ipRange IpRangeData) IsEmpty() bool {
	return ipRange.SyncToken == "" &&
		ipRange.CreateDate == "" &&
		len(ipRange.Prefixes) == 0 &&
		len(ipRange.Ipv6Prefixes) == 0
}

func (IpDataManagerAws *IpDataManagerAws) GetLastModifiedUpstream() (time.Time, error) {
	resp, err := http.Head(IpDataManagerAws.DataURI)
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
func (IpDataManagerAws *IpDataManagerAws) DownloadData() {
	metadataManager := GetMetadataManager()
	resp, err := http.Get(IpDataManagerAws.DataURI)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error downloading dataFile"))
		return
	}

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		util.PrintErrorTrace(util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error downloading dataFile"))
		return
	}

	dataFile, err := os.OpenFile(IpDataManagerAws.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error opening dataFile"))
		return
	}

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error saving dataFile"))
		return
	}

	currentLastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error parsing Date header"))
		return
	}

	err = metadataManager.ReadMetadata()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error reading metadata"))
		return
	}

	if metadataManager.Metadata.LastModified != currentLastModified.Unix() {
		metadata := internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: currentLastModified.Unix(),
		}
		if err := metadataManager.WriteMetadata(&metadata); err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error writing metadata"))
			return
		}
	}

	defer func() {
		if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(networkCloseErr, "Error closing response body"))
		}
		if fileCloseErr := dataFile.Close(); fileCloseErr != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(fileCloseErr, "Error closing dataFile"))
		}
	}()
}

func (IpDataManagerAws *IpDataManagerAws) EnsureDataFile() error {

	metadataManager := GetMetadataManager()
	if !util.IsFileExists(MetadataFilePath) {
		// Create metadata file
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			err = util.ErrorWithInfo(err, "error creating provider directory")
			util.PrintErrorTrace(err)
			return err
		}
		metadataFile, err := os.OpenFile(MetadataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			err = util.ErrorWithInfo(err, "error creating metadata file")
			util.PrintErrorTrace(err)
			return err
		}

		err = metadataManager.WriteMetadata(&internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: 0,
		})

		if err != nil {
			err = util.ErrorWithInfo(err, "error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		defer func() {
			if fileCloseErr := metadataFile.Close(); fileCloseErr != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(fileCloseErr, "error closing metadata file"))
			}
		}()
	}
	if !util.IsFileExists(DataFilePath) {
		// Download the AWS IP ranges file
		IpDataManagerAws.DownloadData()
	}
	if metadataManager.IsExpired() {
		// update the file
		IpDataManagerAws.DownloadData()
	}

	return nil
}

func (IpDataManagerAws *IpDataManagerAws) LoadIpData() *IpRangeData {
	if !IpDataManagerAws.IpRange.IsEmpty() {
		return &IpDataManagerAws.IpRange
	}

	awsIpRangeData := IpRangeData{}
	ipDataFile, err := os.Open(IpDataManagerAws.DataFilePath)
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

	IpDataManagerAws.IpRange = awsIpRangeData
	return &IpDataManagerAws.IpRange
}

var ipDataManagerAws = &IpDataManagerAws{
	DataURI:      DataUrl,
	DataFile:     DataFile,
	DataFilePath: DataFilePath,
	IpRange:      IpRangeData{},
}

func GetIpDataManagerAws() *IpDataManagerAws {
	return ipDataManagerAws
}
