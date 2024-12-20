package gcp

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

func (ipDataManager *IpDataManagerGcp) GetLastModifiedUpstream() (time.Time, error) {
	resp, err := http.Head(ipDataManager.DataURI)
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

func (ipDataManager *IpDataManagerGcp) DownloadData() {
	resp, err := http.Get(ipDataManager.DataURI)
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error downloading dataFile"))
		return
	}

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		util.PrintErrorTrace(util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error downloading dataFile"))
		return
	}

	dataFile, err := os.OpenFile(ipDataManager.DataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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

	err = readMetadata()
	if err != nil {
		util.PrintErrorTrace(util.ErrorWithInfo(err, "Error reading metadata"))
		return
	}

	if metadataManager.Metadata.LastModified != currentLastModified.Unix() {
		metadata := internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: currentLastModified.Unix(),
		}
		if err := writeMetadata(&metadata); err != nil {
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

func (ipDataManager *IpDataManagerGcp) EnsureDataFile() error {
	if !util.IsFileExists(ipDataManager.DataFilePath) {
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			err = util.ErrorWithInfo(err, "Error creating gcp directory")
			util.PrintErrorTrace(err)
			return err
		}
		metadataFile, err := os.OpenFile(MetadataFilePathAws, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			err = util.ErrorWithInfo(err, "Error creating metadata file")
			util.PrintErrorTrace(err)
			return err
		}

		err = writeMetadata(&internal.CloudMetadata{
			Type:         internal.GCP,
			LastModified: 0,
		})

		if err != nil {
			err = util.ErrorWithInfo(err, "Error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		defer func() {
			if fileCloseErr := metadataFile.Close(); fileCloseErr != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(fileCloseErr, "Error closing metadata file"))
			}
		}()
	}
	if !util.IsFileExists(DataFilePathAws) {
		ipDataManagerGcp.DownloadData()
	}
	if isExpired() {
		ipDataManagerGcp.DownloadData()
	}

	return nil
}

func (ipDataManager *IpDataManagerGcp) LoadIpData() *IpRangeDataGcp {
	if !ipDataManager.IpRange.IsEmpty() {
		return &ipDataManagerGcp.IpRange
	}

	gcpIpRangeData := IpRangeDataGcp{}
	ipDataFile, err := os.Open(ipDataManager.DataFilePath)
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

	ipDataManager.IpRange = gcpIpRangeData
	return &ipDataManager.IpRange
}

var ipDataManagerGcp = &IpDataManagerGcp{
	DataURI:      DataUrl,
	DataFile:     DataFile,
	DataFilePath: DataFilePathAws,
	IpRange:      IpRangeDataGcp{},
}
