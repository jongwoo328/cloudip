package ip

import (
	"cloudip/common"
)

var Providers = []common.CloudProvider{
	common.AWS,
	common.GCP,
	common.Azure,
}
