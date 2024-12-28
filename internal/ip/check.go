package ip

import (
	"cloudip/internal"
	"cloudip/internal/ip/aws"
	"cloudip/internal/ip/gcp"
)

func CheckIp(ips *[]string) []internal.CheckIpResult {
	results := make([]internal.CheckIpResult, len(*ips))

	for index, ip := range *ips {
		checkResult, err := checkCloudIp(ip)
		results[index] = internal.CheckIpResult{
			Ip:     ip,
			Result: checkResult,
			Error:  err,
		}
	}

	return results
}

func checkCloudIp(ip string) (internal.Result, error) {
	result := internal.Result{}
	for _, provider := range Providers {
		if provider == internal.AWS {
			isAwsIp, err := aws.IsAwsIp(ip)
			if err != nil {
				return result, err
			}
			if isAwsIp {
				result.Aws = isAwsIp
				return result, nil
			}
		}
		if provider == internal.GCP {
			isGcpIp, err := gcp.IsGcpIp(ip)
			if err != nil {
				return result, err
			}
			if isGcpIp {
				result.Gcp = isGcpIp
				return result, nil
			}
		}
	}
	return result, nil
}
