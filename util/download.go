package util

import (
	"fmt"
	"io"
	"net/http"
)

// DownloadFromUrl reads a complete response using the same bounded client as
// provider JSON downloads. It returns no partial data on interrupted transfers.
func DownloadFromUrl(url string) ([]byte, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return nil, ErrorWithInfo(err, "error downloading data")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading data: received non-200 status code: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorWithInfo(err, "error reading downloaded data")
	}
	return body, nil
}
