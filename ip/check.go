package ip

import (
	"cloudip/common"
	"cloudip/ip/aws"
	"cloudip/ip/gcp"
)

func CheckIp(ips *[]string) []common.CheckIpResult {
	results := make([]common.CheckIpResult, len(*ips))

	for index, ip := range *ips {
		checkResult, err := checkCloudIp(ip)
		results[index] = common.CheckIpResult{
			Ip:     ip,
			Result: checkResult,
			Error:  err,
		}
	}

	return results
}

func checkCloudIp(ip string) (common.Result, error) {
	result := common.Result{}
	for _, provider := range Providers {
		if provider == common.AWS {
			aws.Initialize()
			isAwsIp, err := aws.IsAwsIp(ip)
			if err != nil {
				return result, err
			}
			if isAwsIp {
				result.Aws = isAwsIp
				return result, nil
			}
		}
		if provider == common.GCP {
			gcp.Initialize()
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
