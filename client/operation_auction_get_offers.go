package client

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

type operationAuctionGetOffers struct {
	Category         string   `mapstructure:"1"`
}

func (op operationAuctionGetOffers) Process(state *albionState) {
	log.Debug("Got AuctionGetOffers operation...")
	state.WaitingForMarketData = true
}

type operationAuctionGetOffersResponse struct {
	MarketOrders []string `mapstructure:"0"`
}

func (op operationAuctionGetOffersResponse) Process(state *albionState) {
	log.Debug("Got response to AuctionGetOffers operation...")
	state.WaitingForMarketData = false

	if !state.IsValidLocation() {
		return
	}

	var orders []*lib.MarketOrder

	for _, v := range op.MarketOrders {
		// Unmarshal market order data to map
		var marketOrder map[string]interface{}
		err2 := json.Unmarshal([]byte(v), &marketOrder)
		if err2 != nil {
			log.Fatal(err2)
		}

		// Pull the location
		location, _ := marketOrder["LocationId"].(string)

		// if the location has @, it is either a rest or smugglers den
		if strings.Contains(location, "@") {
			// Remove the @ and everything before it
			location = strings.Split(location, "@")[1]

			// If the location is a smugglers den, remove the BLACKBANK- prefix
			if strings.Contains(location, "BLACKBANK-") {
				location = strings.Replace(location, "BLACKBANK-", "", -1)
			}

			// Set the location in the market order
			marketOrder["LocationId"], _ = strconv.Atoi(location)

			// Marshal the map back to json
			newJson, _ := json.Marshal(marketOrder)

			// Set the new json back to the market order
			v = string(newJson)
		}

		order := &lib.MarketOrder{}

		err := json.Unmarshal([]byte(v), order)
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

	if len(orders) < 1 {
		return
	}

	upload := lib.MarketUpload{
		Orders: orders,
	}

	identifier, _ := uuid.NewV4()
	log.Infof("Sending %d live market sell orders to ingest (Identifier: %s)", len(orders), identifier)
	sendMsgToPublicUploaders(upload, lib.NatsMarketOrdersIngest, state, identifier.String())
}
