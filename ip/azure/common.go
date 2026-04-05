package azure

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"sync"
)

var appDir = util.GetAppDir(common.AppName)

const DataFile = "azure.json"
const MetadataFile = ".metadata.json"

var dataRequestOnce sync.Once
var dataUrl string = "" // return empty string when error

func getDataUrl() string {
	dataRequestOnce.Do(func() {

		resp, err := http.Get("https://www.microsoft.com/en-us/download/details.aspx?id=56519")
		if err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "error checking metadata file expiration"))
			return
		}

		if resp.StatusCode != http.StatusOK {
			err := util.ErrorWithInfo(fmt.Errorf("received non-200 status code: %s", resp.Status), "error checking metadata file expiration")
			util.PrintErrorTrace(err)
			return
		}

		defer func() {
			if networkCloseErr := resp.Body.Close(); networkCloseErr != nil {
				util.PrintErrorTrace(util.ErrorWithInfo(networkCloseErr, "error closing response body"))
			}
		}()

		document, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			util.PrintErrorTrace(util.ErrorWithInfo(err, "error parsing html"))
			return
		}

		downloadButton := document.Find(`section[aria-label="download action"] a`).First()
		if downloadButton == nil {
			util.PrintErrorTrace(util.ErrorWithInfo(errors.New("no download button found"), "error parsing html"))
			return
		}

		href, exists := downloadButton.Attr("href")
		if !exists {
			util.PrintErrorTrace(util.ErrorWithInfo(errors.New("no href attribute found"), "error parsing html"))
			return
		}

		dataUrl = href
	})

	return dataUrl
}

var ProviderDirectory = fmt.Sprintf("%s/%s", appDir, "azure")
var DataFilePathAzure = fmt.Sprintf("%s/%s", ProviderDirectory, DataFile)
var MetadataFilePathAzure = fmt.Sprintf("%s/%s", ProviderDirectory, MetadataFile)
