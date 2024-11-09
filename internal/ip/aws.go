package ip

import (
	"cloudip/internal"
	"cloudip/internal/ip/aws"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

var v4Tree *CIDRTree
var v6Tree *CIDRTree

type AwsIpRangeData struct {
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

type AwsDataMetadata struct {
	LastModified int64 `json:"lastModified"`
}

func init() {
	v4Tree = NewCIDRTree()
	v6Tree = NewCIDRTree()

	ensureAwsIpFile()
	awsIpRangeData := AwsIpRangeData{}
	err := getAwsData(&awsIpRangeData)
	if err != nil {
		fmt.Println("Error getting AWS IP range data:", err)
		return
	}

	for _, prefix := range awsIpRangeData.Prefixes {
		v4Tree.AddCIDR(prefix.IpPrefix)
	}

	for _, prefix := range awsIpRangeData.Ipv6Prefixes {
		v6Tree.AddCIDR(prefix.Ipv6Prefix)
	}
}

func IsAwsIp(ip string) (bool, error) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return false, fmt.Errorf("Error parsing IP: %s", ip)
	}

	if parsedIp.To4() != nil {
		return v4Tree.Match(ip), nil
	}
	if parsedIp.To16() != nil {
		return v6Tree.Match(ip), nil
	}
	return false, nil
}

func ensureAwsIpFile() {
	metadataFilePath := fmt.Sprintf("%s/%s", getAwsDataDir(), aws.MetadataFile)
	if !internal.IsFileExists(metadataFilePath) {
		// Create metadata file
		metadataFile, err := os.OpenFile(metadataFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("Error creating metadata file:", err)
			return
		}
		err = writeMetadata(metadataFile, &AwsDataMetadata{
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
	if !internal.IsFileExists(aws.DataFilePath) {
		// Download the AWS IP ranges file
		downloadData()
	} else if isExpired() {
		// update the file
		downloadData()
	}
}

func getAwsDataDir() string {
	appDir, err := internal.GetAppDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	awsDir := fmt.Sprintf("%s/aws", appDir)
	internal.EnsureDir(awsDir)
	return awsDir
}

func downloadData() {
	resp, err := http.Get(aws.DataUrl)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer func() {
		if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
			fmt.Println("Error closing response body:", networkCloseErr)
		}
	}()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Received non-200 status code:", resp.Status)
		return
	}

	// 파일 생성
	file, err := os.Create(aws.DataFilePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer func() {
		if fileCloseErr := file.Close(); fileCloseErr != nil {
			fmt.Println("Error closing file:", fileCloseErr)
		}
	}()

	// 응답 본문을 파일로 복사
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	currentLastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		fmt.Println("Error parsing Date header:", err)
		return
	}

	var metadataFile *os.File
	metadataFile, err = os.OpenFile(aws.MetadataFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening metadata file:", err)
		return
	}
	defer func() {
		if fileCloseErr := metadataFile.Close(); fileCloseErr != nil {
			fmt.Println("Error closing metadata file:", fileCloseErr)
		}
	}()

	storedMetadata := AwsDataMetadata{}
	err = getMetadata(metadataFile, &storedMetadata)
	if err != nil {
		fmt.Println("Error getting metadata:", err)
		return
	}

	if storedMetadata.LastModified != currentLastModified.Unix() {
		metadata := AwsDataMetadata{
			LastModified: currentLastModified.Unix(),
		}
		if err := writeMetadata(metadataFile, &metadata); err != nil {
			fmt.Println("Error writing metadata:", err)
			return
		}
	}
}

func isExpired() bool {
	metadataFile, err := os.Open(aws.MetadataFilePath)
	if err != nil {
		fmt.Println("Error opening MetadataFile:", err)
		return true
	}
	defer func() {
		if fileCloseErr := metadataFile.Close(); fileCloseErr != nil {
			fmt.Println("Error closing MetadataFile:", fileCloseErr)
		}
	}()

	resp, err := http.Head(aws.DataUrl)
	if err != nil {
		fmt.Println("Error checking MetadataFile expiration:", err)
		return false
	}
	defer func() {
		if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
			fmt.Println("Error closing response body:", networkCloseErr)
		}
	}()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Received non-200 status code:", resp.Status)
		return false
	}

	// Last-Modified 헤더 확인
	lastModified := resp.Header.Get("Last-Modified")
	lastModifiedDate, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		fmt.Println("Error parsing Date header:", err)
		return false
	}

	metadata := AwsDataMetadata{}
	err = getMetadata(metadataFile, &metadata)
	if err != nil {
		fmt.Println("Error getting metadata:", err)
		return false
	}

	return lastModifiedDate.Unix() != metadata.LastModified
}

func getMetadata[T any](metadataFile *os.File, metadata *T) error {
	// check file existance
	_, err := metadataFile.Stat()
	if err != nil {
		return err
	}
	return internal.HandleJSON(metadataFile, metadata, "read")
}

func writeMetadata[T any](metadataFile *os.File, metadata *T) error {
	if _, err := metadataFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek the file: %v", err)
	}
	return internal.HandleJSON(metadataFile, metadata, "write")
}

func getAwsData(data *AwsIpRangeData) error {
	file, err := os.Open(aws.DataFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer func() {
		if fileCloseErr := file.Close(); fileCloseErr != nil {
			fmt.Println("Error closing file:", fileCloseErr)
		}
	}()

	err = internal.HandleJSON(file, data, "read")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}

	return nil
}
