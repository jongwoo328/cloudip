package cmd

import (
	"cloudip/internal"
	"cloudip/internal/ip"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

type CloudIpArgs struct {
	pretty    bool
	delimiter string
	version   bool
}

var Args = CloudIpArgs{}
var Version string

var rootCmd = &cobra.Command{
	Use:   internal.AppName,
	Short: fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", internal.AppName),
	Run: func(cmd *cobra.Command, args []string) {
		if Args.version {
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
	rootCmd.Flags().BoolVar(&Args.pretty, "pretty", false, "Pretty print the output")
	rootCmd.Flags().StringVar(&Args.delimiter, "delimiter", " ", "Delimiter for the output")
	rootCmd.Flags().BoolVarP(&Args.version, "version", "v", false, "Print the version")
}

func getProviderFromResult(result internal.Result) string {
	if result.Aws {
		return "aws"
	}
	if result.Gcp {
		return "gcp"
	}
	return "unknown"
}

func printResult(results *[]internal.CheckIpResult) {
	if Args.pretty {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"IP", "Provider"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		for _, r := range *results {
			var provider string
			if r.Error != nil {
				provider = "ERROR"
			} else {
				provider = getProviderFromResult(r.Result)
			}
			table.Append([]string{r.Ip, provider})
		}
		table.Render()
	} else {
		for _, r := range *results {
			var provider string
			if r.Error != nil {
				provider = "ERROR"
			} else {
				provider = getProviderFromResult(r.Result)
			}
			fmt.Printf("%s%s%s\n", r.Ip, Args.delimiter, provider)
		}
	}

}
