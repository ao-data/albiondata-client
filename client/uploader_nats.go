package client

import (
	"github.com/ao-data/albiondata-client/log"
	nats "github.com/nats-io/go-nats"
)

type natsUploader struct {
	isPrivate bool
	url       string
	nc        *nats.Conn
}

// newNATSUploader creates a new NATS uploader
func newNATSUploader(url string) uploader {
	nc, _ := nats.Connect(url)

	return &natsUploader{
		url: url,
		nc:  nc,
	}
}

func (u *natsUploader) sendToIngest(body []byte, topic string, state *albionState, identifier string) {
	// not handling sending identifier since the official usage is with http_pow

	if err := u.nc.Publish(topic, body); err != nil {
		log.Errorf("Error while sending ingest to nats with data: %v", err)
	}
}
