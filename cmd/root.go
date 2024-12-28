package cmd

import (
	"cloudip/internal"
	"cloudip/internal/ip"
	"fmt"
	"github.com/spf13/cobra"
)

type CloudIpFlag struct {
	delimiter string
	version   bool
	format    string
	header    bool
}

var Flags = CloudIpFlag{}
var Version string

var rootCmd = &cobra.Command{
	Use:   internal.AppName,
	Short: fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", internal.AppName),
	Run: func(cmd *cobra.Command, args []string) {
		if Flags.version {
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
	rootCmd.Flags().BoolVarP(&Flags.version, "version", "v", false, "Print the version")

	rootCmd.Flags().StringVarP(&Flags.format, "format", "f", "text", "Output format (text, table, json)")
	rootCmd.Flags().BoolVar(&Flags.header, "header", false, "Print header in the output. Only applicable for 'text', 'table' format")
	rootCmd.Flags().StringVar(&Flags.delimiter, "delimiter", " ", "Delimiter for the output. Only applicable for 'text' format")
}
