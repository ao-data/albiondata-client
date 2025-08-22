package client

import (
	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationAuctionBuyOfferRequest struct {
	Amount                   int8  `mapstructure:"1"`
	ID                       int64 `mapstructure:"2"`
	TransferItemsToInventory bool  `mapstructure:"3"`
}

func (op operationAuctionBuyOfferRequest) Process(state *albionState) {
	log.Debug("Got AuctionBuyOfferRequest operation...")

	if !state.IsValidLocation() {
		return
	}

	upload := lib.MarketBuyItemRequest{
		SellOrderID: int(op.ID),
		Amount:      int(op.Amount),
		LocationID:  state.LocationId,
	}

	identifier, _ := uuid.NewV4()
	log.Infof("Sending market buy request to ingest (Identifier: %s)", identifier)
	sendMsgToPublicUploaders(upload, lib.NatsMarketBoughtOrdersIngest, state, identifier.String())
}
