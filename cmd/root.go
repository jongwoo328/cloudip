package cmd

import (
	"cloudip/internal"
	"cloudip/internal/ip"
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   internal.AppName,
	Short: fmt.Sprintf("%s is a CLI tool for identifying whether an IP address belongs to a major cloud provider (e.g., AWS, GCP).", internal.AppName),
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("인수 목록:")
		for i, arg := range args {
			fmt.Printf("  인수 %d: %s\n", i+1, arg)
		}
		result := ip.CheckIp(&args)
		for _, r := range result {
			fmt.Println(r)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {

	// 서브커맨드 추가
	//rootCmd.AddCommand(subCmd)
	// 플래그 추가
	//rootCmd.Flags().StringP("package", "p", "", "Specify the package to install")
}
