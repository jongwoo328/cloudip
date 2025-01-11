package util

import "net/http"

func GetHeadRequestHeader(url string) (http.Header, error) {
	resp, err := http.Head(url)
	if err != nil {
		return nil, ErrorWithInfo(err, "Error getting head request header")
	}
	if err := resp.Body.Close(); err != nil {
		return nil, ErrorWithInfo(err, "Error closing response body")
	}

	return resp.Header, nil
}
