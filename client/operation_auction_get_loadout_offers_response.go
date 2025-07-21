package client

import (
	"encoding/json"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationAuctionGetLoadoutOffersResponse struct {
	MarketOrders [][]string `mapstructure:"1"`
	Quantities   [][]int    `mapstructure:"2"`
}

func (op operationAuctionGetLoadoutOffersResponse) Process(state *albionState) {
	log.Infof("Got response to AuctionGetOffers operation...")

	if !state.IsValidLocation() {
		return
	}

	var orders []*lib.MarketOrder

	// For loadouts the MarketOrders are grouped by item.
	// For example if we have food in our loadout and we want 10 of it, and there are 2 sell orders one with 6 and other with 4,
	// we will get 2 entries in one of the MarketOrders array entries and the Quantities array will have an entrywith 6 and 4 aswell.
	for _, groupedOrders := range op.MarketOrders {
		for _, unparsedOrder := range groupedOrders {
			// Unmarshal market order data to map
			var marketOrder map[string]interface{}
			err2 := json.Unmarshal([]byte(unparsedOrder), &marketOrder)
			if err2 != nil {
				log.Fatal(err2)
			}

			order := &lib.MarketOrder{}

			err := json.Unmarshal([]byte(unparsedOrder), order)
			if err != nil {
				log.Errorf("Problem converting market order to internal struct: %v", err)
			}

			// Set the location only if its 0. Smugglers Dens pull locations directly from the market data (above)
			// while the orignal cities have a null location ID and is pulled from the client state.
			if order.LocationID == 0 {
				order.LocationID = state.LocationId
			}

			orders = append(orders, order)
		}
	}

	if len(orders) < 1 {
		return
	}

	// upload := lib.MarketUpload{
	// 	Orders: orders,
	// }

	identifier, _ := uuid.NewV4()
	log.Infof("Sending %d loadout sell orders to ingest (Identifier: %s)", len(orders), identifier)
	for _, order := range orders {
		log.Infof("Order: %+v", order)
	}
	//sendMsgToPublicUploaders(upload, lib.NatsMarketOrdersIngest, state, identifier.String())
}
