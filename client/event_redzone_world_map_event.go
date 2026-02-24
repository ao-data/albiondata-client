package client

import (
	"time"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

/*
The event is received when the world map is opened or dragged and sometimes randomly?.

EventDataType: [475]evRedZoneWorldMapEvent - map[0:639075060010075967 1:1 252:475]
EventDataType: [475]evRedZoneWorldMapEvent - map[0:639075087010222801 1:2 252:475]
EventDataType: [475]evRedZoneWorldMapEvent - map[0:639075105010297587 1:3 2:[DUCHY_RED_01 DUCHY_RED_05] 252:475]

map[0] - Timestamp when the phase ends
map[1] - Current phase (1-3)
map[2] - References the provinces in the 3rd phase
*/

type eventRedZoneWorldMapEvent struct {
	EventTime     int64 `mapstructure:"0"`
	Phase         int   `mapstructure:"1"`
}

func (event eventRedZoneWorldMapEvent) Process(state *albionState) {
	log.Debug("Got red zone world event...")

	if state.BanditEventLastTimeSubmitted.IsZero() || time.Since(state.BanditEventLastTimeSubmitted).Seconds() >= 60 {
		state.BanditEventLastTimeSubmitted = time.Now()

		log.Infof("Bandit Event detected (Phase: %d) ending at %d", event.Phase, event.EventTime)

		identifier, _ := uuid.NewV4()
		upload := lib.BanditEvent{
			EventTime:     event.EventTime,
			Phase:         event.Phase,
		}
		log.Infof("Sending bandit event to ingest (Identifier: %s)", identifier)
		sendMsgToPublicUploaders(upload, lib.NatsBanditEvent, state, identifier.String())
	}
}
