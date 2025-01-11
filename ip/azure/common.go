package azure

import (
	"cloudip/util"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"sync"
)

var appDir = util.GetAppDir()

const DataFile = "azure.json"
const MetadataFile = ".metadata.json"

var dataRequestOnce sync.Once
var dataUrl string = "" // return empty string when error

func getDataUrl() string {
	dataRequestOnce.Do(func() {

		resp, err := http.Get("https://www.microsoft.com/en-us/download/details.aspx?id=56519")
		if err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error checking metadata file expiration"))
			return
		}

		if resp.StatusCode != http.StatusOK {
			err := util.ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error checking metadata file expiration")
			util.PrintErrorTrace(err)
			return
		}

		defer func() {
			if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(networkCloseErr, "Error closing response body"))
			}
		}()

		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "Error parsing html"))
			return
		}

		downloadButton := document.Find(`section[aria-label="download action"] a`).First()
		if downloadButton == nil {
			util.PrintErrorTrace(util.ErrorWithInfo(errors.New("No download button found"), "Error parsing html"))
			return
		}

		href, exists := downloadButton.Attr("href")
		if !exists {
			util.PrintErrorTrace(util.ErrorWithInfo(errors.New("No href attribute found"), "Error parsing html"))
			return
		}

		dataUrl = href
	})

	return dataUrl
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "azure")
var DataFilePathAzure = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathAzure = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
