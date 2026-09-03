package client

import "testing"

// TestNewHTTPUploader_SetsRequestTimeout guards against a request hanging
// forever when the network stalls (e.g. mid connectivity issue). Combined
// with router.go spawning one goroutine per decoded operation, a client
// with no timeout meant a stalled upload pinned that goroutine - and
// everything it was holding - in memory indefinitely.
func TestNewHTTPUploader_SetsRequestTimeout(t *testing.T) {
	u, ok := newHTTPUploader("https://example.com").(*httpUploader)
	if !ok {
		t.Fatalf("newHTTPUploader did not return a *httpUploader")
	}
	if u.client.Timeout <= 0 {
		t.Fatalf("httpUploader.client.Timeout = %v, want a positive timeout", u.client.Timeout)
	}
}

// TestNewHTTPUploaderPow_SetsRequestTimeout is the http+pow equivalent of
// TestNewHTTPUploader_SetsRequestTimeout - it made two unbounded HTTP
// calls per upload (fetch the pow challenge, then submit the solution).
func TestNewHTTPUploaderPow_SetsRequestTimeout(t *testing.T) {
	u, ok := newHTTPUploaderPow("https+pow://example.com").(*httpUploaderPow)
	if !ok {
		t.Fatalf("newHTTPUploaderPow did not return a *httpUploaderPow")
	}
	if u.client.Timeout <= 0 {
		t.Fatalf("httpUploaderPow.client.Timeout = %v, want a positive timeout", u.client.Timeout)
	}
}
