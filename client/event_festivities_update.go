package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

const csharpTicksUnixEpoch int64 = 621355968000000000

type eventFestivitiesUpdate struct {
	Kinds       []uint8  `mapstructure:"0"`
	Categories  []string `mapstructure:"1"`
	UniqueNames []string `mapstructure:"2"`
	StartTimes  []int64  `mapstructure:"3"`
	EndTimes    []int64  `mapstructure:"4"`
}

func (event eventFestivitiesUpdate) Process(state *albionState) {
	log.Debug("Got festivities update event...")

	if !state.FestivitiesLastTimeSubmitted.IsZero() && time.Since(state.FestivitiesLastTimeSubmitted).Seconds() < 60 {
		return
	}

	upload, err := event.upload()
	if err != nil {
		log.Errorf("Could not parse festivities update: %v", err)
		return
	}

	state.FestivitiesLastTimeSubmitted = time.Now()

	identifier, _ := uuid.NewV4()
	log.Infof("Sending %d festivities to ingest (Identifier: %s)", len(upload.Events), identifier)
	sendMsgToPublicUploaders(upload, lib.NatsFestivitiesIngest, state, identifier.String(), len(upload.Events))
}

func (event eventFestivitiesUpdate) upload() (lib.FestivitiesUpload, error) {
	count := len(event.UniqueNames)
	if count == 0 || len(event.Kinds) != count || len(event.Categories) != count || len(event.StartTimes) != count || len(event.EndTimes) != count {
		return lib.FestivitiesUpload{}, fmt.Errorf(
			"inconsistent array lengths: kinds=%d categories=%d names=%d starts=%d ends=%d",
			len(event.Kinds), len(event.Categories), count, len(event.StartTimes), len(event.EndTimes),
		)
	}

	upload := lib.FestivitiesUpload{
		Events: make([]*lib.Festivity, 0, count),
	}

	for i := 0; i < count; i++ {
		uniqueName := strings.TrimSpace(event.UniqueNames[i])
		if uniqueName == "" {
			return lib.FestivitiesUpload{}, fmt.Errorf("empty unique name at index %d", i)
		}
		if event.StartTimes[i] < csharpTicksUnixEpoch || event.EndTimes[i] <= event.StartTimes[i] {
			return lib.FestivitiesUpload{}, fmt.Errorf("invalid event interval at index %d", i)
		}

		upload.Events = append(upload.Events, &lib.Festivity{
			Kind:       event.Kinds[i],
			Category:   event.Categories[i],
			UniqueName: uniqueName,
			StartTime:  event.StartTimes[i],
			EndTime:    event.EndTimes[i],
		})
	}

	return upload, nil
}
