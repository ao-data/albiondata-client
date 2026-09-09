package client

import (
	"encoding/json"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationAuctionGetLoadoutOffersResponse struct {
	LoadoutOrders     [][]string `mapstructure:"1"`
	LoadoutQuantities [][]int    `mapstructure:"2"`
}

func (op operationAuctionGetLoadoutOffersResponse) Process(state *albionState) {
	if !state.IsValidLocation() {
		return
	}

	var orders = parseLoadoutOrders(op.LoadoutOrders, state)

	if len(orders) < 1 {
		return
	}

	uploadOrders(orders, state)
}

func uploadOrders(orders []*lib.MarketOrder, state *albionState) {
	identifier, _ := uuid.NewV4()
	log.Infof("Sending %d loadout sell orders to ingest (Identifier: %s)", len(orders), identifier)
	upload := lib.MarketUpload{
		Orders: orders,
	}
	sendMsgToPublicUploaders(upload, lib.NatsMarketOrdersIngest, state, identifier.String())
}

func parseLoadoutOrders(loadoutOrders [][]string, state *albionState) []*lib.MarketOrder {
	var orders []*lib.MarketOrder

	// For loadouts the MarketOrders are grouped by item.
	// For example if we have food in our loadout and we want 10 of it, and there are 2 sell orders one with 6 and other with 4,
	// we will get 2 entries in one of the MarketOrders array entries and the Quantities array will have an entrywith 6 and 4 aswell.
	for _, loadoutOrder := range loadoutOrders {
		for _, unparsedOrder := range loadoutOrder {
			order := &lib.MarketOrder{}
			unmarshalOrder(unparsedOrder, order)

			// Loadout orders are always from the current location.
			if order.LocationID == 0 {
				order.LocationID = state.LocationId
			}

			orders = append(orders, order)
		}
	}

	return orders
}

func unmarshalOrder(unparsedOrder string, parsedOrder *lib.MarketOrder) {
	err := json.Unmarshal([]byte(unparsedOrder), parsedOrder)
	if err != nil {
		log.Errorf("Problem converting market order to internal struct: %v", err)
	}
}
