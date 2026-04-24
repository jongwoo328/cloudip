package util

import (
	"fmt"
	"net/http"
	"time"
)

var headClient = &http.Client{
	Timeout: 5 * time.Second,
}

func GetHeadRequestHeader(url string) (http.Header, error) {
	resp, err := headClient.Head(url)
	if err != nil {
		return nil, ErrorWithInfo(err, "error getting head request header")
	}
	if err := resp.Body.Close(); err != nil {
		return nil, ErrorWithInfo(err, "error closing response body")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrorWithInfo(fmt.Errorf("received non-200 status code: %s", resp.Status), "error getting head request header")
	}

	return resp.Header, nil
}
