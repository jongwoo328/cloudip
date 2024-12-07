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
}

var Args = CloudIpArgs{}

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
	rootCmd.Flags().BoolVar(&Args.pretty, "pretty", false, "Pretty print the output")
	rootCmd.Flags().StringVar(&Args.delimiter, "delimiter", " ", "Delimiter for the output")
}

func getProviderFromResult(result internal.Result) string {
	if result.Aws {
		return "aws"
	}
	return "unknown"
}

func printResult(results *[]internal.CheckIpResult) {
	if Args.pretty {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"IP", "Provider"})
		for _, r := range *results {
			var provider = getProviderFromResult(r.Result)
			table.Append([]string{r.Ip, provider})
		}
		table.Render()
	} else {
		for _, r := range *results {
			var provider = getProviderFromResult(r.Result)
			fmt.Printf("%s%s%s\n", r.Ip, Args.delimiter, provider)
		}
	}

}
