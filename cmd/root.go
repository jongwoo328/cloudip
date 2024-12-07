package cmd

import (
	"cloudip/internal"
	"cloudip/internal/ip"
	"fmt"
	"github.com/spf13/cobra"
)

type CloudIpArgs struct {
	pretty bool
}

var Args = CloudIpArgs{pretty: false}

var rootCmd = &cobra.Command{
	Use:   internal.AppName,
	Short: fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", internal.AppName),
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		result := ip.CheckIp(&args)
		printResult(&result)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	//rootCmd.Flags().StringP("package", "p", "", "Specify the package to install")
	rootCmd.Flags().BoolVar(&Args.pretty, "pretty", false, "Pretty print the output")
}

func printResult(results *[]internal.CheckIpResult) {
	if Args.pretty {
		// Pretty print the output
	} else {
		for _, r := range *results {
			var provider string
			if r.Result.Aws {
				provider = "aws"
			} else {
				provider = "unknown"
			}
			fmt.Printf("%s\t%s\n", r.Ip, provider)
		}
	}

}
