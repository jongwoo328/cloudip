package aws

import (
	"cloudip/internal"
	"cloudip/internal/util"
	"fmt"
	"os"
)

var appDir, _ = util.GetAppDir()

const DataFile = "aws.json"
const MetadataFile = ".metadata.json"
const DataUrl = "https://ip-ranges.amazonaws.com/ip-ranges.json"

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "aws")
var DataFilePath = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePath = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)

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

func EnsureAwsIpFile() {
	metadataManager := GetMetadataManager()
	if !util.IsFileExists(MetadataFilePath) {
		// Create metadata file
		if err := os.MkdirAll(ProviderDirectory, 0755); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		metadataFile, err := os.OpenFile(MetadataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("Error creating metadata file:", err)
			return
		}

		err = metadataManager.WriteMetadata(&internal.CloudMetadata{
			Type:         internal.AWS,
			LastModified: 0,
		})

		if err != nil {
			fmt.Println("Error writing metadata:", err)
			return
		}
		defer func() {
			if fileCloseErr := metadataFile.Close(); fileCloseErr != nil {
				fmt.Println("Error closing metadata file:", fileCloseErr)
			}
		}()
	}
	if !util.IsFileExists(DataFilePath) {
		// Download the AWS IP ranges file
		metadataManager.DownloadData()
		return
	}
	if metadataManager.IsExpired() {
		// update the file
		metadataManager.DownloadData()
	}

}
