package client

import (
	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type eventRedZonePlayerNotification struct{}

func (event eventRedZonePlayerNotification) Process(state *albionState) {
	log.Debug("Got red zone player notification...")

	if !ConfigGlobal.NotifyBanditEvents {
		return
	}

	identifier, _ := uuid.NewV4()
	upload := lib.BanditsEventNotification{}
	sendMsgToPrivateUploaders(&upload, "bandits_event_notification", state, identifier.String())
}
