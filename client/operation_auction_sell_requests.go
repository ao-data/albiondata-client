package client

import (
	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationAuctionSellSpecificItemRequest struct {
	ID        int64 `mapstructure:"1"`
	ItemIndex int16 `mapstructure:"2"`
	// T         int8  `mapstructure:"3` Not sure what this value is (Low number item specific)
	Amount int8 `mapstructure:"4"`
}

func (op operationAuctionSellSpecificItemRequest) Process(state *albionState) {
	log.Debug("Got AuctionSellSpecificItemRequest operation...")

	if !state.IsValidLocation() {
		return
	}

	upload := lib.MarketSellItemRequest{
		BuyOrderID: int(op.ID),
		ItemIndex:  int(op.ItemIndex),
		Amount:     int(op.Amount),
		LocationID: state.LocationId,
	}

	identifier, _ := uuid.NewV4()
	log.Infof("Sending market sell request to ingest (Identifier: %s)", identifier)
	sendMsgToPublicUploaders(upload, lib.NatsMarketSoldOrdersIngest, state, identifier.String())
}
