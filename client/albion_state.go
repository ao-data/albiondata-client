package client

import (
	"github.com/broderickhyman/albiondata-client/lib"
	"github.com/broderickhyman/albiondata-client/log"
	"github.com/broderickhyman/albiondata-client/notification"

	"strings"
)

//CacheSize limit size of messages in cache
const CacheSize = 256

type marketHistoryInfo struct {
	albionId  uint32
	timescale lib.Timescale
	quality   uint8
}

type albionState struct {
	LocationId     int
	LocationString string
	CharacterId    lib.CharacterID
	CharacterName  string
	GameServerIP   string
	AODataServerID int

	// A lot of information is sent out but not contained in the response when requesting marketHistory (e.g. ID)
	// This information is stored in marketHistoryInfo
	// This array acts as a type of cache for that info
	// The index is the message number (param255) % CacheSize
	marketHistoryIDLookup [CacheSize]marketHistoryInfo
	// TODO could this be improved?!
}

func (state albionState) IsValidLocation() bool {
	if state.LocationId < 0 {
		if state.LocationId == -1 {
			log.Error("The players location has not yet been set. Please transition zones so the location can be identified.")
			if !ConfigGlobal.Debug {
				notification.Push("The players location has not yet been set. Please transition zones so the location can be identified.")
			}
		} else {
			log.Error("The players location is not valid. Please transition zones so the location can be fixed.")
			if !ConfigGlobal.Debug {
				notification.Push("The players location is not valid. Please transition zones so the location can be fixed.")
			}
		}
		return false
	}
	return true
}

func (state albionState) SetServerID() int {
	if state.AODataServerID == 0 {
		if strings.HasPrefix(state.GameServerIP, "5.188.125.") {
			log.Error("Set Albion Data Project Server ID to 1")
			return 1
		} else if strings.HasPrefix(state.GameServerIP, " 5.45.187.") {
			log.Error("Set Albion Data Project Server ID to 2")
			return 2
		}
	} else {
		return state.AODataServerID
	}

	return 0
}
