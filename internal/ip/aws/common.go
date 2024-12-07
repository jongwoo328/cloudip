package aws

import (
	"cloudip/internal/util"
	"fmt"
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
