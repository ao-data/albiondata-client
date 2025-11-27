package client

import (
	"fmt"
	"time"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/notification"
	uuid "github.com/nu7hatch/gouuid"
)

/*
DEBU[2025-11-26T11:30:45Z] Unhandled event type: evRedZoneWorldEvent
DEBU[2025-11-26T11:30:45Z] EventDataType: [474]evRedZoneWorldEvent - map[0:638997538711438026 1:true 252:474]
*/
type eventRedZoneWorldEvent struct {
	EventTime int64 `mapstructure:"0"`
	Arg1      bool  `mapstructure:"1"`
}

func (event eventRedZoneWorldEvent) Process(state *albionState) {
	log.Debug("Got red zone world event...")

	if !ConfigGlobal.NotifyBanditEvents {
		return
	}

	if state.BanditEventStartTime != event.EventTime {
		state.BanditEventStartTime = event.EventTime

		// convert .net ticks to formated time
		eventTime := NetTicksToTime(event.EventTime).Format("2006-01-02 15:04:05")
		message := fmt.Sprintf("World Event at %s", eventTime)
		log.Info("Bandit/World Event detected")
		notification.Push(message)

		identifier, _ := uuid.NewV4()
		upload := lib.BanditsEventNotification{
			EventTime: eventTime,
		}
		sendMsgToPrivateUploaders(&upload, "bandits_event_notification", state, identifier.String())
	}
}

func NetTicksToTime(netTicks int64) time.Time {
	const ticksToUnixEpoch int64 = 621355968000000000

	// Get ticks elapsed since 1970
	unixTicks := netTicks - ticksToUnixEpoch

	// Convert to seconds (10,000,000 ticks per second)
	seconds := unixTicks / 10_000_000

	// Take the remainder and convert to nanoseconds (1 tick = 100 ns)
	nanos := (unixTicks % 10_000_000) * 100

	return time.Unix(seconds, nanos).UTC()
}
