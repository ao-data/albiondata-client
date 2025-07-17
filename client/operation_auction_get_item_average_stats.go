package client

import (
	"sort"
	"time"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationAuctionGetItemAverageStats struct {
	ItemID      int32         `mapstructure:"1"`
	Quality     uint8         `mapstructure:"2"`
	Timescale   lib.Timescale `mapstructure:"3"`
	Enchantment uint32        `mapstructure:"4"`
	MessageID   uint64        `mapstructure:"255"`
}

func (op operationAuctionGetItemAverageStats) Process(state *albionState) {
	var index = op.MessageID % CacheSize

	// It seems all items with id 129-256 come through as a negative integer. Example, goose eggs
	// comes through as -121. (-121)+256=135. As of today (2024-01-07), the itemId in the ao-bin-dumps repo
	// is 135. This occurs for all items we can search the market for with english text from id 128-256.
	// Anything 128 and below or 256 and greater seem to work just fine. - phendryx 2024-01-07
	var itemId = op.ItemID
	if itemId < 0 && itemId > -129 {
		itemId = itemId + 256
	} else {
		itemId = op.ItemID
	}

	mhInfo := marketHistoryInfo{
		albionId:  itemId,
		timescale: op.Timescale,
		quality:   op.Quality,
	}

	state.marketHistoryIDLookup[index] = mhInfo
	log.Debugf("Market History - Caching %d at %d.", mhInfo.albionId, index)
}

type operationAuctionGetItemAverageStatsResponse struct {
	ItemAmounts   []int64  `mapstructure:"0"`
	SilverAmounts []uint64 `mapstructure:"1"`
	Timestamps    []uint64 `mapstructure:"2"`
	MessageID     int      `mapstructure:"255"`
}

func (op operationAuctionGetItemAverageStatsResponse) Process(state *albionState) {
	var index = op.MessageID % CacheSize

	// Wait for the correlating Request if it has not yet been processed
	waits := 0
	for waits < 30 {
		if state.marketHistoryIDLookup[index].albionId < 1 {
			time.Sleep(1 * time.Second)
			waits += 1
		} else {
			break
		}
	}

	// Still no correlating Request has been processed
	if state.marketHistoryIDLookup[index].albionId < 1 {
		log.Warnf("Market History - Market history at index %d is invalid. Has albionId: %s ", index, state.marketHistoryIDLookup[index].albionId)
		return
	}

	var mhInfo = state.marketHistoryIDLookup[index]

	// Clear the index in the cache
	state.marketHistoryIDLookup[index].albionId = 0
	log.Debugf("Market History - Loaded itemID %d from cache at index %d", mhInfo.albionId, index)
	log.Debug("Got response to GetItemAverageStats operation for the itemID[", mhInfo.albionId, "] of quality: ", mhInfo.quality, " and on the timescale: ", mhInfo.timescale)

	if !state.IsValidLocation() {
		return
	}

	var histories []*lib.MarketHistory

	// TODO can we make this safer? Right now we just assume all the arrays are the same length as the number of item amounts
	for i := range op.ItemAmounts {
		// sometimes opAuctionGetItemAverageStats receives negative item amounts
		if op.ItemAmounts[i] < 0 {
			if op.ItemAmounts[i] < -124 {
				// still don't know what to do with these
				log.Debugf("Market History - Ignoring negative item amount %d for %d silver on %d", op.ItemAmounts[i], op.SilverAmounts[i], op.Timestamps[i])
				continue
			}
			// however these can be interpreted by adding them to 256
			// TODO: make more sense of this, (perhaps there is a better way)
			log.Debugf("Market History - Interpreting negative item amount %d as %d for %d silver on %d", op.ItemAmounts[i], 256+op.ItemAmounts[i], op.SilverAmounts[i], op.Timestamps[i])
			op.ItemAmounts[i] = 256 + op.ItemAmounts[i]
		}
		history := &lib.MarketHistory{}
		history.ItemAmount = op.ItemAmounts[i]
		history.SilverAmount = op.SilverAmounts[i]
		history.Timestamp = op.Timestamps[i]
		histories = append(histories, history)
	}

	if len(histories) < 1 {
		log.Info("Auction Stats Response - no history\n\n")
		return
	}

	// Sort history by descending time so the newest is always first in the list
	sort.SliceStable(histories, func(i, j int) bool {
		return histories[i].Timestamp > histories[j].Timestamp
	})

	upload := lib.MarketHistoriesUpload{
		AlbionId:     mhInfo.albionId,
		LocationId:   state.LocationId,
		QualityLevel: mhInfo.quality,
		Timescale:    mhInfo.timescale,
		Histories:    histories,
	}

	identifier, _ := uuid.NewV4()
	log.Infof("Sending %d market history item average stats to ingest for albionID %d (Identifier: %s)", len(histories), mhInfo.albionId, identifier)
	sendMsgToPublicUploaders(upload, lib.NatsMarketHistoriesIngest, state, identifier.String())
}
