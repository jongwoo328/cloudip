package cmd

import (
	"cloudip/common"
	"cloudip/ip"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
)

var Version string

func NewRootCmd(flags *common.CloudIpFlag, checker *ip.IPChecker) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           common.AppName,
		Short:         fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", common.AppName),
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			common.SetVerbose(flags.Verbose)
			policy := common.DefaultUpdatePolicy()
			policy.NoUpdate = flags.NoUpdate
			result := checker.Check(args, policy)
			if err := printResult(cmd.OutOrStdout(), result, flags); err != nil {
				return err
			}
			if !hasResultError(result) {
				return nil
			}
			printResultErrors(cmd.ErrOrStderr(), result)
			return errors.New("one or more IP checks failed")
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
	rootCmd.Flags().BoolVar(&flags.NoUpdate, "no-update", false, "Use local provider data without checking for updates")
	rootCmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Print verbose output")

	return rootCmd
}

func hasResultError(results []common.Result) bool {
	for _, result := range results {
		if result.Error != nil {
			return true
		}
	}
	return false
}

func printResultErrors(w io.Writer, results []common.Result) {
	for _, result := range results {
		if result.Error == nil {
			continue
		}
		fmt.Fprintf(w, "%s: %v\n", result.Ip, result.Error)
	}
}
