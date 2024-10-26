package ip

type CheckIpResult struct {
	Ip     string
	Result Result
}

type Result struct {
	Aws bool
}

var providers = []string{
	"aws",
}

func CheckIp(ips *[]string) []CheckIpResult {
	results := make([]CheckIpResult, len(*ips))
	for index, ip := range *ips {
		results[index] = CheckIpResult{
			Ip:     ip,
			Result: checkCloudIp(ip),
		}
	}
	return results
}

func checkCloudIp(ip string) Result {
	result := Result{}
	for _, provider := range providers {
		if provider == "aws" {
			isAwsIp, err := IsAwsIp(ip)
			if err != nil {
				continue
			}
			result.Aws = isAwsIp
		}
	}
	return result
}
