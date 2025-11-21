package client

import (
	"fmt"

	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/notification"
)

type eventRedZoneWorldEvent struct {
	EventType int    `mapstructure:"0"`
	EventID   string `mapstructure:"1"`
	Location  string `mapstructure:"2"`
}

func (event eventRedZoneWorldEvent) Process(state *albionState) {
	log.Debug("Got red zone world event...")

	if !ConfigGlobal.NotifyBanditEvents {
		return
	}

	// Event types:
	// Bandit events typically have specific event type identifiers
	// We'll notify on all world events for now, but can be filtered if needed
	var message string
	if event.EventID != "" {
		message = fmt.Sprintf("World Event [%s] at: %s", event.EventID, event.Location)
	} else {
		message = fmt.Sprintf("World Event Detected: %s", event.Location)
	}

	log.Infof("Bandit/World Event detected: %v at %v", event.EventID, event.Location)
	notification.Push(message)
}
