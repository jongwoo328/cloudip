package aws

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

func (ipDataManagerAws *IpDataManagerAws) GetSignatureUpstream() (string, error) {
	headers, err := util.GetHeadRequestHeader(ipDataManagerAws.DataURI)
	if err != nil {
		err = util.ErrorWithInfo(err, "error getting header from request")
		util.PrintErrorTrace(err)
		return "", err
	}

	signature, _, err := awsSignatureFromHeaders(headers)
	return signature, err
}

func (ipDataManagerAws *IpDataManagerAws) downloadData() error {
	common.VerboseOutput("Downloading AWS IP ranges...")
	if ipDataManagerAws.DataURI == "" {
		return errors.New("cannot get DataURI")
	}

	headers, err := util.DownloadFromUrlToFileWithHeaders(ipDataManagerAws.DataURI, ipDataManagerAws.DataFilePath)
	if err != nil {
		return err
	}

	signature, currentLastModified, err := awsSignatureFromHeaders(headers)
	if err != nil {
		util.PrintErrorTrace(err)
		return err
	}

	if metadataManager.IsSignatureExpired(signature) {
		metadata := common.CloudMetadata{
			Type:      common.AWS,
			Signature: signature,
		}
		if err := metadataManager.Write(&metadata); err != nil {
			err = util.ErrorWithInfo(err, "error writing metadata")
			util.PrintErrorTrace(err)
			return err
		}
		common.VerboseOutput(fmt.Sprintf("AWS IP ranges updated [%s]", util.FormatToTimestamp(currentLastModified)))
	}

	return nil
}

func awsSignatureFromHeaders(headers http.Header) (string, time.Time, error) {
	currentLastModified, err := time.Parse(time.RFC1123, headers.Get("Last-Modified"))
	if err != nil {
		return "", time.Time{}, util.ErrorWithInfo(err, "error parsing Date header")
	}

	signature := strings.TrimSpace(headers.Get("ETag"))
	signature = strings.TrimPrefix(signature, "W/")
	signature = strings.Trim(signature, "\"")
	if signature == "" {
		signature = common.LastModifiedSignature(currentLastModified)
	}
	return signature, currentLastModified, nil
}

func (ipDataManagerAws *IpDataManagerAws) EnsureDataFile() error {
	if err := metadataManager.Ensure(); err != nil {
		return err
	}
	if err := metadataManager.Read(); err != nil {
		return err
	}

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
		err = util.ErrorWithInfo(err, "error opening data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}
	defer ipDataFile.Close()

	err = util.ReadJSON(ipDataFile, &awsIpRangeData)
	if err != nil {
		err = util.ErrorWithInfo(err, "error reading data file")
		util.PrintErrorTrace(err)
		log.Fatal(err)
	}

	ipDataManagerAws.IpRange = awsIpRangeData
	return &ipDataManagerAws.IpRange
}

var ipDataManagerAws = &IpDataManagerAws{
	DataURI:      getDataUrl(),
	DataFile:     DataFile,
	DataFilePath: DataFilePathAws,
	IpRange:      IpRangeDataAws{},
}
