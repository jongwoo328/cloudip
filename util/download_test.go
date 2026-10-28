package util

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type bytesTransport func(*http.Request) (*http.Response, error)

func (f bytesTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type downloadBody struct {
	io.Reader
	closed  bool
	failure error
}

func (b *downloadBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	if err == io.EOF && b.failure != nil {
		return n, b.failure
	}
	return n, err
}
func (b *downloadBody) Close() error { b.closed = true; return nil }

func TestDownloadBytesRejectsPartialResponses(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		readErr error
		wantErr bool
	}{
		{"complete response", http.StatusOK, nil, false},
		{"non-200 response", http.StatusServiceUnavailable, nil, true},
		{"interrupted body", http.StatusOK, io.ErrUnexpectedEOF, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := downloadClient
			t.Cleanup(func() { downloadClient = original })
			body := &downloadBody{Reader: strings.NewReader("192.0.2.0/24,,,,\n"), failure: tc.readErr}
			downloadClient = &http.Client{Timeout: original.Timeout, Transport: bytesTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Status: http.StatusText(tc.status), Body: body, Header: make(http.Header), Request: r}, nil
			})}
			data, err := DownloadFromUrl("https://example.invalid/feed.csv")
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v", err)
			}
			if tc.wantErr && data != nil {
				t.Fatal("returned partial or unsuccessful response data")
			}
			if !tc.wantErr && string(data) != "192.0.2.0/24,,,,\n" {
				t.Fatalf("data = %q", data)
			}
			if tc.readErr != nil && !errors.Is(err, tc.readErr) {
				t.Fatalf("read error lost: %v", err)
			}
			if !body.closed {
				t.Fatal("response body was not closed")
			}
		})
	}
}

func TestDownloadBytesPreservesTransportError(t *testing.T) {
	original := downloadClient
	t.Cleanup(func() { downloadClient = original })
	failure := errors.New("transport unavailable")
	downloadClient = &http.Client{Timeout: original.Timeout, Transport: bytesTransport(func(*http.Request) (*http.Response, error) { return nil, failure })}
	data, err := DownloadFromUrl("https://example.invalid/feed.csv")
	if data != nil || !errors.Is(err, failure) {
		t.Fatalf("data = %q, error = %v", data, err)
	}
}
