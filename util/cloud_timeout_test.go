package util

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadTimeoutIncludesStalledResponseBody(t *testing.T) {
	if downloadClient.Timeout <= 0 {
		t.Fatal("downloads must have an overall timeout")
	}
	original := downloadClient
	client := *original
	client.Timeout = 100 * time.Millisecond
	downloadClient = &client
	t.Cleanup(func() { downloadClient = original })
	for _, kind := range []string{"file", "JSON", "bytes"} {
		t.Run(kind, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{"))
				w.(http.Flusher).Flush()
				select {
				case <-r.Context().Done():
				case <-time.After(3 * time.Second):
				}
			}))
			defer server.Close()
			start := time.Now()
			var err error
			if kind == "file" {
				_, err = DownloadFromUrlToFileWithHeaders(server.URL, filepath.Join(t.TempDir(), "data.json"))
			} else if kind == "bytes" {
				_, err = DownloadFromUrl(server.URL)
			} else {
				var data map[string]any
				_, err = DownloadJSONFromUrl(server.URL, &data)
			}
			var timeout net.Error
			if !errors.As(err, &timeout) || !timeout.Timeout() {
				t.Fatalf("expected body timeout, got %v", err)
			}
			if elapsed := time.Since(start); elapsed > 2*time.Second {
				t.Fatalf("stalled body took %s", elapsed)
			}
		})
	}
}
