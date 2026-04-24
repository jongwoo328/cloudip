package util

import (
	"fmt"
	"io"
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

// GetRequestHeader는 GET 응답 헤더를 반환한다. 본문은 버린다.
// Cloudflare의 CDN edge처럼 HEAD와 GET 응답 헤더가 다른 경우에 사용.
func GetRequestHeader(url string) (http.Header, error) {
	resp, err := headClient.Get(url)
	if err != nil {
		return nil, ErrorWithInfo(err, "error getting request header")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrorWithInfo(fmt.Errorf("received non-200 status code: %s", resp.Status), "error getting request header")
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return nil, ErrorWithInfo(err, "error discarding response body")
	}
	return resp.Header, nil
}
