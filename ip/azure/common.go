package azure

import (
	"cloudip/util" // Keep for PrintErrorTrace and ErrorWithInfo in getDataUrl
	"errors"
	"fmt" // Keep for Errorf in getDataUrl
	"github.com/PuerkitoBio/goquery"
	"github.com/ip-api/cloudip/ip" // Added for ip.Get... functions
	"net/http"
	"sync"
)

// var appDir = util.GetAppDir() // Removed

const DataFile = "azure.json" // Keep: specific to Azure
// const MetadataFile = ".metadata.json" // Removed: ip.GetMetadataFilePath uses ip.DefaultMetadataFile

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

var ProviderDirectory = ip.GetProviderDirectory("azure")
var DataFilePathAzure = ip.GetDataFilePath("azure", DataFile)
var MetadataFilePathAzure = ip.GetMetadataFilePath("azure")
