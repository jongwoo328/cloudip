package cmd

import (
	"cloudip/common"
	"cloudip/ip"
	"fmt"
	"github.com/spf13/cobra"
)

var Version string

var rootCmd = &cobra.Command{
	Use:   common.AppName,
	Short: fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", common.AppName),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] == "version" {
			fmt.Println(Version)
			return
		}
		result := ip.CheckIp(&args)
		printResult(&result)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&common.Flags.Format, "format", "f", "text", "Output format (text, table, json)")
	rootCmd.Flags().BoolVar(&common.Flags.Header, "header", false, "Print header in the output. Only applicable for 'text', 'table' format")
	rootCmd.Flags().StringVar(&common.Flags.Delimiter, "delimiter", " ", "Delimiter for the output. Only applicable for 'text' format")
	rootCmd.Flags().BoolVarP(&common.Flags.Verbose, "verbose", "v", false, "Print verbose output")
}
