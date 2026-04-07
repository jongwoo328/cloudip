package cmd

import (
	"cloudip/common"
	"cloudip/ip"
	"fmt"
	"github.com/spf13/cobra"
)

var Version string

func NewRootCmd(flags *common.CloudIpFlag, checker *ip.IPChecker) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   common.AppName,
		Short: fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", common.AppName),
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			common.SetVerbose(flags.Verbose)
			result := checker.Check(args)
			printResult(&result, flags)
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of cloudip",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().StringVarP(&flags.Format, "format", "f", "text", "Output format (text, table, json)")
	rootCmd.Flags().BoolVar(&flags.Header, "header", false, "Print header in the output. Only applicable for 'text' format")
	rootCmd.Flags().StringVar(&flags.Delimiter, "delimiter", " ", "Delimiter for the output. Applicable for 'text' and 'table' format")
	rootCmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Print verbose output")

	return rootCmd
}
