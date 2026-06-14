package cmd

import (
	"cloudip/common"
	"encoding/json"
	"fmt"
	"io"

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

type jsonResult struct {
	IP       string `json:"ip"`
	Provider string `json:"provider"`
	Error    string `json:"error"`
}

func printResult(w io.Writer, results []common.Result, flags *common.CloudIpFlag) error {
	switch flags.Format {
	case "text":
		return printResultAsText(w, results, flags)
	case "table":
		return printResultAsTable(w, results, flags)
	case "json":
		return printResultAsJson(w, results)
	default:
		return fmt.Errorf("invalid output format: %s. Supported formats are: text, table, json", flags.Format)
	}
}

func printResultAsText(w io.Writer, results []common.Result, flags *common.CloudIpFlag) error {
	if flags.Header {
		if _, err := fmt.Fprintf(w, "%s%s%s\n", headers["IP"], flags.Delimiter, headers["Provider"]); err != nil {
			return fmt.Errorf("error writing text result: %w", err)
		}
	}
	for _, r := range results {
		if _, err := fmt.Fprintf(w, "%s%s%s\n", r.Ip, flags.Delimiter, getProviderString(r)); err != nil {
			return fmt.Errorf("error writing text result: %w", err)
		}
	}
	return nil
}
func printResultAsTable(w io.Writer, results []common.Result, flags *common.CloudIpFlag) error {
	table := tablewriter.NewTable(w,
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
	for _, r := range results {
		table.Append(r.Ip, getProviderString(r))
	}
	if err := table.Render(); err != nil {
		return fmt.Errorf("error writing table result: %w", err)
	}
	return nil
}

func printResultAsJson(w io.Writer, results []common.Result) error {
	resultSlice := make([]jsonResult, 0, len(results))
	for _, r := range results {
		result := jsonResult{
			IP:       r.Ip,
			Provider: getJSONProviderString(r),
			Error:    getErrorString(r),
		}
		resultSlice = append(resultSlice, result)
	}
	bytes, err := json.Marshal(resultSlice)
	if err != nil {
		return fmt.Errorf("error converting result to JSON: %w", err)
	}
	if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
		return fmt.Errorf("error writing JSON result: %w", err)
	}
	return nil
}

func getJSONProviderString(r common.Result) string {
	if r.Error != nil {
		return "error"
	}
	return getProviderString(r)
}

func getErrorString(r common.Result) string {
	if r.Error == nil {
		return ""
	}
	return r.Error.Error()
}
