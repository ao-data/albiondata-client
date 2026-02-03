package client
import (
	"time"
	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/lib"
	uuid "github.com/nu7hatch/gouuid"
)
/*
The event is received when the world map is opened or dragged and sometimes randomly?.

The event can either be
> EventDataType: [474]evRedZoneWorldEvent - map[0:638997538711438026 1:true 252:474]
  * Happens 15 minutes before the event actually starts.
  * Gives the timestamp of when the event starts

_OR_

> EventDataType: [474]evRedZoneWorldEvent - map[0:639054601760934861 252:474]
  * Happens when the event has already started.
  * Gives the timestamp of when the event ends
*/

type eventRedZoneWorldEvent struct {
	EventTime int64 `mapstructure:"0"`
	AdvanceNotice bool `mapstructure:"1"`
}

func (event eventRedZoneWorldEvent) Process(state *albionState) {
	log.Debug("Got red zone world event...")

	if state.BanditEventLastTimeSubmitted.IsZero() || time.Since(state.BanditEventLastTimeSubmitted).Seconds() >= 60 {
		state.BanditEventLastTimeSubmitted = time.Now()

		if event.AdvanceNotice {
			log.Infof("Bandit Event detected starting at %d", event.EventTime)
		} else {
			log.Infof("Bandit Event detected ending at %d", event.EventTime)
		}
		
		identifier, _ := uuid.NewV4()
		upload := lib.BanditEvent{
			EventTime: event.EventTime,
			AdvanceNotice: event.AdvanceNotice,
		}
		log.Infof("Sending bandit event to ingest (Identifier: %s)", identifier)
		sendMsgToPublicUploaders(upload, lib.NatsBanditEvent, state, identifier.String())
	}
}
