package azure

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type IpDataManagerAzure struct {
	DataURI      string
	DataFile     string
	DataFilePath string
	IpRange      IpRangeDataAzure
	dataURIMu    sync.Mutex
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

func (ipDataManagerAzure *IpDataManagerAzure) ensureDataURI() error {
	ipDataManagerAzure.dataURIMu.Lock()
	defer ipDataManagerAzure.dataURIMu.Unlock()

	if ipDataManagerAzure.DataURI != "" {
		return nil
	}

	dataURI := getDataUrl()
	if dataURI == "" {
		return errors.New("cannot get DataURI")
	}

	ipDataManagerAzure.DataURI = dataURI
	return nil
}

func (ipRange IpRangeDataAzure) IsEmpty() bool {
	return ipRange.ChangeNumber == 0 &&
		len(ipRange.Values) == 0
}

func (ipDataManagerAzure *IpDataManagerAzure) GetLastModifiedUpstream() (time.Time, error) {
	if err := ipDataManagerAzure.ensureDataURI(); err != nil {
		return time.Time{}, err
	}

	headers, err := util.GetHeadRequestHeader(ipDataManagerAzure.DataURI)
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

func (ipDataManagerAzure *IpDataManagerAzure) downloadData() error {
	common.VerboseOutput("Downloading Azure IP ranges...")

	if err := ipDataManagerAzure.ensureDataURI(); err != nil {
		return err
	}

	headers, err := util.DownloadFromUrlToFileWithHeaders(ipDataManagerAzure.DataURI, ipDataManagerAzure.DataFilePath)
	if err != nil {
		return err
	}

	currentLastModified, err := time.Parse(time.RFC1123, headers.Get("Last-Modified"))
	if err != nil {
		err = util.ErrorWithInfo(err, "error parsing Date header")
		util.PrintErrorTrace(err)
		return err
	}
	signature := common.LastModifiedSignature(currentLastModified)

	if metadataManager.IsSignatureExpired(signature) {
		metadata := common.CloudMetadata{
			Type:      common.Azure,
			Signature: signature,
		}
		if err := metadataManager.Write(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput(fmt.Sprintf("Azure IP ranges updated [%s]", util.FormatToTimestamp(currentLastModified)))

	}

	return nil
}

func (ipDataManagerAzure *IpDataManagerAzure) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

	if !util.IsFileExists(DataFilePathAzure) {
		common.VerboseOutput("Azure IP ranged file not exists.")
		err := ipDataManagerAzure.downloadData()
		return err
	}
	if isExpired() {
		common.VerboseOutput("Azure IP ranged are outdated. Updating to the latest version...")
		err := ipDataManagerAzure.downloadData()
		return err
	}
	common.VerboseOutput("Azure IP ranged are up-to-date.")

	return nil
}

func (ipDataManagerAzure *IpDataManagerAzure) LoadIpData() (*IpRangeDataAzure, error) {
	if !ipDataManagerAzure.IpRange.IsEmpty() {
		return &ipDataManagerAzure.IpRange, nil
	}

	azureIpRangeData := IpRangeDataAzure{}
	ipDataFile, err := os.Open(ipDataManagerAzure.DataFilePath)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error loading data file")
	}
	defer ipDataFile.Close()

	err = util.ReadJSON(ipDataFile, &azureIpRangeData)
	if err != nil {
		return nil, util.ErrorWithInfo(err, "error reading data file")
	}

	ipDataManagerAzure.IpRange = azureIpRangeData
	return &ipDataManagerAzure.IpRange, nil
}

var ipDataManagerAzure = &IpDataManagerAzure{
	DataFile:     DataFile,
	DataFilePath: DataFilePathAzure,
	IpRange:      IpRangeDataAzure{},
}
