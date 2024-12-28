package cmd

import (
	"cloudip/internal"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

func getProviderFromResult(result internal.Result) string {
	if result.Aws {
		return "aws"
	}
	if result.Gcp {
		return "gcp"
	}
	return "unknown"
}

var headers = map[string]string{
	"IP":       "IP",
	"Provider": "Provider",
}

func printResult(results *[]internal.CheckIpResult) {
	if Flags.format == "text" {
		printResultAsText(results)
	} else if Flags.format == "table" {
		printResultAsTable(results)
	} else if Flags.format == "json" {
		printResultAsJson(results)
	} else {
		fmt.Printf("Invalid output format: %s. Supported formats are: text, table, json\n", Flags.format)
	}
}

func printResultAsText(results *[]internal.CheckIpResult) {
	if Flags.header {
		fmt.Printf("%s%s%s\n", headers["IP"], Flags.delimiter, headers["Provider"])
	}
	for _, r := range *results {
		var provider string
		if r.Error != nil {
			provider = "ERROR"
		} else {
			provider = getProviderFromResult(r.Result)
		}
		fmt.Printf("%s%s%s\n", r.Ip, Flags.delimiter, provider)
	}
}
func printResultAsTable(results *[]internal.CheckIpResult) {
	table := tablewriter.NewWriter(os.Stdout)
	if Flags.header {
		table.SetHeader([]string{headers["IP"], headers["Provider"]})
	}
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	tableData := make([][]string, 0)
	for _, r := range *results {
		var provider string
		if r.Error != nil {
			provider = "ERROR"
		} else {
			provider = getProviderFromResult(r.Result)
		}
		tableData = append(tableData, []string{r.Ip, provider})
	}
	table.AppendBulk(tableData)
	table.Render()
}

func printResultAsJson(results *[]internal.CheckIpResult) {
	resultSlice := make([]map[string]string, 0)
	for _, r := range *results {
		resultMap := map[string]string{
			headers["IP"]:       r.Ip,
			headers["Provider"]: getProviderFromResult(r.Result),
		}
		resultSlice = append(resultSlice, resultMap)
	}
	bytes, err := json.Marshal(resultSlice)
	if err != nil {
		fmt.Println("Error converting result to JSON")
		os.Exit(1)
	}
	fmt.Println(string(bytes))
}
