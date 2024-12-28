package ip

import (
	"cloudip/internal"
	"cloudip/internal/ip/aws"
	"cloudip/internal/ip/gcp"
	"sync"
)

func CheckIp(ips *[]string) []internal.CheckIpResult {
	results := make([]internal.CheckIpResult, len(*ips))
	var waitGroup sync.WaitGroup

	for index, ip := range *ips {
		waitGroup.Add(1)
		go func(index int, ip string) {
			defer waitGroup.Done()
			checkResult, err := checkCloudIp(ip)
			results[index] = internal.CheckIpResult{
				Ip:     ip,
				Result: checkResult,
				Error:  err,
			}
		}(index, ip)
	}

	waitGroup.Wait()
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
			result.Aws = isAwsIp
		} else if provider == internal.GCP {
			isGcpIp, err := gcp.IsGcpIp(ip)
			if err != nil {
				return result, err
			}
			result.Gcp = isGcpIp
		}
	}
	return result, nil
}
