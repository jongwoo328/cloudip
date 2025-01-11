package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFromUrlToFile(url string, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "Error downloading data file"))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := ErrorWithInfo(fmt.Errorf("Received non-200 status code: %s", resp.Status), "Error downloading data file")
		PrintErrorTrace(err)
		return err
	}
	dataFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		err = ErrorWithInfo(err, "Error opening data file")
		PrintErrorTrace(err)
		return err
	}

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		err = ErrorWithInfo(err, "Error saving data file")
		PrintErrorTrace(err)
		return err
	}

	defer func() {
		if networkCloseEre := resp.Body.Close(); networkCloseEre != nil {
			PrintErrorTrace(ErrorWithInfo(networkCloseEre, "Error closing response body"))
		}
		if fileCloseErr := dataFile.Close(); fileCloseErr != nil {
			PrintErrorTrace(ErrorWithInfo(fileCloseErr, "Error closing data file"))
		}
	}()

	return nil
}
