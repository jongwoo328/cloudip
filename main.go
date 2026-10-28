package main

import (
	"cloudip/cmd"
	"cloudip/common"
	"cloudip/ip"
	"cloudip/ip/aws"
	"cloudip/ip/azure"
	"cloudip/ip/cloudflare"
	"cloudip/ip/digitalocean"
	"cloudip/ip/gcp"
	"cloudip/ip/provider"
	"cloudip/util"
	"os"
)

func main() {
	if err := util.EnsureAppDir(common.AppName); err != nil {
		util.PrintErrorTrace(err)
		os.Exit(1)
	}

	flags := &common.CloudIpFlag{}
	checker := ip.NewIPChecker(
		map[common.CloudProvider]provider.CloudProvider{
			common.AWS:          aws.Provider,
			common.GCP:          gcp.Provider,
			common.Azure:        azure.Provider,
			common.Cloudflare:   cloudflare.Provider,
			common.DigitalOcean: digitalocean.Provider,
		},
		ip.DefaultProviderOrder,
	)

	if err := cmd.NewRootCmd(flags, checker).Execute(); err != nil {
		util.PrintErrorTrace(err)
		os.Exit(1)
	}
}
