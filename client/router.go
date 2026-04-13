package client

import (
	"encoding/gob"
	"os"

	"github.com/ao-data/albiondata-client/client/photon"
	"github.com/ao-data/albiondata-client/log"
)

// Router struct definitions
type Router struct {
	albionstate     *albionState
	newOperation    chan operation
	recordRawPacket chan photon.RawPacket
	quit            chan bool
}

func newRouter() *Router {
	return &Router{
		albionstate:     &albionState{LocationId: ""},
		newOperation:    make(chan operation, 1000),
		recordRawPacket: make(chan photon.RawPacket, 1000),
		quit:            make(chan bool, 1),
	}
}

func (r *Router) run() {
	var encoder *gob.Encoder
	var file *os.File
	if ConfigGlobal.RecordPath != "" {
		file, err := os.Create(ConfigGlobal.RecordPath)
		if err != nil {
			log.Error("Could not open commands output file ", err)
		} else {
			encoder = gob.NewEncoder(file)
		}
	}

	for {
		select {
		case <-r.quit:
			log.Debug("Closing router...")
			if file != nil {
				err := file.Close()
				if err != nil {
					log.Error("Could not close commands output file ", err)
				}
			}
			return
		case op := <-r.newOperation:
			go op.Process(r.albionstate)
		case raw := <-r.recordRawPacket:
			if encoder != nil {
				err := encoder.Encode(raw)
				if err != nil {
					log.Error("Could not encode raw packet ", err)
				}
			}
		}
	}
}
