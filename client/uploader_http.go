package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// httpUploadTimeout bounds every ingest HTTP request. Without it, a
// stalled connection during a network hiccup blocks the request's
// goroutine (one is spawned per decoded operation, see router.go)
// forever instead of failing and letting that goroutine's memory go.
const httpUploadTimeout = 30 * time.Second

type httpUploader struct {
	baseURL string
	client  *http.Client
}

// newHTTPUploader creates a new HTTP uploader
func newHTTPUploader(url string) uploader {
	return &httpUploader{
		baseURL: url,
		client:  &http.Client{Transport: &http.Transport{}, Timeout: httpUploadTimeout},
	}
}

func (u *httpUploader) sendToIngest(body []byte, topic string, state *albionState, identifier string) {
	// not handling sending identifier since the official usage is with http_pow

	fullURL := u.baseURL + "/" + topic

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		log.Errorf("Error while create new request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := u.client.Do(req)
	if err != nil {
		log.Errorf("Error while sending ingest with data: %v", err)
		return
	}

	if resp.StatusCode != 200 {
		log.Errorf("Got bad response code: %v", resp.StatusCode)
		return
	}

	// See: https://stackoverflow.com/questions/17948827/reusing-http-connections-in-golang
	io.Copy(ioutil.Discard, resp.Body)

	log.Infof("Successfully sent ingest request to %v", u.baseURL)

	defer resp.Body.Close()
}
