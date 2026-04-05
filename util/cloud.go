package util

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

var downloadClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	},
}

func DownloadFromUrlToFile(url string, filePath string) error {
	resp, err := downloadClient.Get(url)
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "error downloading data file"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := ErrorWithInfo(fmt.Errorf("received non-200 status code: %s", resp.Status), "error downloading data file")
		PrintErrorTrace(err)
		return err
	}

	dataFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		err = ErrorWithInfo(err, "error opening data file")
		PrintErrorTrace(err)
		return err
	}
	defer dataFile.Close()

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		err = ErrorWithInfo(err, "error saving data file")
		PrintErrorTrace(err)
		return err
	}

	return nil
}
