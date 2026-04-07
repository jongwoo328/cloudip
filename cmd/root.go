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
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if common.Flags.Format == "table" && !cmd.Flags().Changed("delimiter") {
			common.Flags.Delimiter = "\t"
		}
		result := ip.Check(args)
		printResult(&result)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of cloudip",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(Version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().StringVarP(&common.Flags.Format, "format", "f", "text", "Output format (text, table, json)")
	rootCmd.Flags().BoolVar(&common.Flags.Header, "header", false, "Print header in the output. Only applicable for 'text' format")
	rootCmd.Flags().StringVar(&common.Flags.Delimiter, "delimiter", " ", "Delimiter for the output (default: space for text, tab for table)")
	rootCmd.Flags().BoolVarP(&common.Flags.Verbose, "verbose", "v", false, "Print verbose output")
}
