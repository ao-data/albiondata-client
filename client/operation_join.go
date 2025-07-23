package client

import (
	"strconv"
	"strings"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
)

type operationJoinResponse struct {
	CharacterID   lib.CharacterID `mapstructure:"1"`
	CharacterName string          `mapstructure:"2"`
	Location      string          `mapstructure:"8"`
	GuildID       lib.CharacterID `mapstructure:"53"`
	GuildName     string          `mapstructure:"57"`
}

//CharacterPartsJSON string          `mapstructure:"6"`
//Edition            string          `mapstructure:"38"`

func (op operationJoinResponse) Process(state *albionState) {
	log.Debugf("Got JoinResponse operation...")

	// Reset the AODataServerID here. This leads to a fresh execution
	// of SetServerID() incase the player switched servers
	state.AODataServerID = 0

	// Hack for second caerleon marketplace
	if strings.HasSuffix(op.Location, "-Auction2") {
		op.Location = strings.Replace(op.Location, "-Auction2", "", -1)
	}

	// Allow for smugglers rest locations markets to be parsed by setting a valid location int
	if strings.HasPrefix(op.Location, "BLACKBANK-") {
		op.Location = strings.Replace(op.Location, "BLACKBANK-", "", -1)
	}

	// Abyssal Depths Patch 2
	// Antiquarian’s Dens now contain a Marketplace vendor for the corresponding city’s Marketplace
	if strings.HasSuffix(op.Location, "-HellDen") {
		op.Location = strings.Replace(op.Location, "-HellDen", "", -1)
	}

	loc, err := strconv.Atoi(op.Location)
	if err != nil {
		log.Debugf("Unable to convert zoneID to int. Probably an instance.")
		state.LocationId = -2
	} else {
		state.LocationId = loc
	}
	log.Infof("Updating player location to %v.", op.Location)

	if state.CharacterId != op.CharacterID {
		log.Infof("Updating player ID to %v.", op.CharacterID)
	}
	state.CharacterId = op.CharacterID

	if state.CharacterName != op.CharacterName {
		log.Infof("Updating player to %v.", op.CharacterName)
	}
	state.CharacterName = op.CharacterName
}
