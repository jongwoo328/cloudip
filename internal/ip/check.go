package ip

import (
	"cloudip/internal"
	"cloudip/internal/ip/aws"
	"cloudip/internal/ip/gcp"
)

func CheckIp(ips *[]string) []internal.CheckIpResult {
	results := make([]internal.CheckIpResult, len(*ips))
	for index, ip := range *ips {
		results[index] = internal.CheckIpResult{
			Ip:     ip,
			Result: checkCloudIp(ip),
		}
	}
	return results
}

func checkCloudIp(ip string) internal.Result {
	result := internal.Result{}
	for _, provider := range Providers {
		if provider == internal.AWS {
			isAwsIp, err := aws.IsAwsIp(ip)
			if err != nil {
				continue
			}
			result.Aws = isAwsIp
		} else if provider == internal.GCP {
			isGcpIp, err := gcp.IsGcpIp(ip)
			if err != nil {
				continue
			}
			result.Gcp = isGcpIp
		}
	}
	return result
}
