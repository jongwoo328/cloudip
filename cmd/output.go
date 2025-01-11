package cmd

import (
	"cloudip/common"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
)

func getProviderFromResult(result common.Result) string {
	if result.Aws {
		return "aws"
	}
	if result.Gcp {
		return "gcp"
	}
	if result.Azure {
		return "azure"
	}
	return "unknown"
}

var headers = map[string]string{
	"IP":       "IP",
	"Provider": "Provider",
}

func printResult(results *[]common.CheckIpResult) {
	if common.Flags.Format == "text" {
		printResultAsText(results)
	} else if common.Flags.Format == "table" {
		printResultAsTable(results)
	} else if common.Flags.Format == "json" {
		printResultAsJson(results)
	} else {
		fmt.Printf("Invalid output format: %s. Supported formats are: text, table, json\n", common.Flags.Format)
	}
}

func printResultAsText(results *[]common.CheckIpResult) {
	if common.Flags.Header {
		fmt.Printf("%s%s%s\n", headers["IP"], common.Flags.Delimiter, headers["Provider"])
	}
	for _, r := range *results {
		var provider string
		if r.Error != nil {
			provider = "ERROR"
		} else {
			provider = getProviderFromResult(r.Result)
		}
		fmt.Printf("%s%s%s\n", r.Ip, common.Flags.Delimiter, provider)
	}
}
func printResultAsTable(results *[]common.CheckIpResult) {
	table := tablewriter.NewWriter(os.Stdout)
	if common.Flags.Header {
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

func printResultAsJson(results *[]common.CheckIpResult) {
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
