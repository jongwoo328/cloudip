package ip

import (
	"cloudip/common"
)

var DefaultProviderOrder = []common.CloudProvider{
	common.AWS,
	common.GCP,
	common.Azure,
}
