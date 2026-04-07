package cmd

import (
	"cloudip/common"
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func getProviderString(r common.Result) string {
	if r.Error != nil {
		return "ERROR"
	}
	if r.Provider == "" {
		return "unknown"
	}
	return string(r.Provider)
}

var headers = map[string]string{
	"IP":       "IP",
	"Provider": "Provider",
}

func printResult(results *[]common.Result, flags *common.CloudIpFlag) error {
	switch flags.Format {
	case "text":
		printResultAsText(results, flags)
	case "table":
		printResultAsTable(results, flags)
	case "json":
		printResultAsJson(results)
	default:
		return fmt.Errorf("invalid output format: %s. Supported formats are: text, table, json", flags.Format)
	}
	return nil
}

func printResultAsText(results *[]common.Result, flags *common.CloudIpFlag) {
	if flags.Header {
		fmt.Printf("%s%s%s\n", headers["IP"], flags.Delimiter, headers["Provider"])
	}
	for _, r := range *results {
		fmt.Printf("%s%s%s\n", r.Ip, flags.Delimiter, getProviderString(r))
	}
}
func printResultAsTable(results *[]common.Result, flags *common.CloudIpFlag) {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders:  tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			Symbols:  tw.NewSymbolCustom("delim").WithColumn(flags.Delimiter),
			Settings: tw.Settings{Lines: tw.LinesNone},
		})),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithHeaderAutoFormat(tw.Off),
		tablewriter.WithPadding(tw.PaddingNone),
	)

	table.Header(headers["IP"], headers["Provider"])
	for _, r := range *results {
		table.Append(r.Ip, getProviderString(r))
	}
	table.Render()
}

func printResultAsJson(results *[]common.Result) {
	resultSlice := make([]map[string]string, 0)
	for _, r := range *results {
		resultMap := map[string]string{
			headers["IP"]:       r.Ip,
			headers["Provider"]: getProviderString(r),
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
