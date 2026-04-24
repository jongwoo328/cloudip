package util

import (
	"encoding/json"
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
	_, err := DownloadFromUrlToFileWithHeaders(url, filePath)
	return err
}

func DownloadFromUrlToFileWithHeaders(url string, filePath string) (http.Header, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "error downloading data file"))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := ErrorWithInfo(fmt.Errorf("received non-200 status code: %s", resp.Status), "error downloading data file")
		PrintErrorTrace(err)
		return nil, err
	}

	dataFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		err = ErrorWithInfo(err, "error opening data file")
		PrintErrorTrace(err)
		return nil, err
	}
	defer dataFile.Close()

	_, err = io.Copy(dataFile, resp.Body)
	if err != nil {
		err = ErrorWithInfo(err, "error saving data file")
		PrintErrorTrace(err)
		return nil, err
	}

	return resp.Header, nil
}

func DownloadJSONFromUrl[T any](url string, data *T) (http.Header, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		PrintErrorTrace(ErrorWithInfo(err, "error downloading JSON data"))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := ErrorWithInfo(fmt.Errorf("received non-200 status code: %s", resp.Status), "error downloading JSON data")
		PrintErrorTrace(err)
		return nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		err = ErrorWithInfo(err, "error decoding JSON data")
		PrintErrorTrace(err)
		return nil, err
	}

	return resp.Header, nil
}
