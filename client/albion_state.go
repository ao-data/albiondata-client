package client

import (
	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/notification"
	"strings"
)

//CacheSize limit size of messages in cache
const CacheSize = 8192

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

func (state albionState) GetServerID() int {
	// default to 0
	var serverID = 0

	// if we happen to have a server id stored in state, lets re-default to that
	if state.AODataServerID != 0 {
		serverID = state.AODataServerID
	}

	// we get packets from other than game servers, so determine if it's a game server
	// based on soruce ip and if its east/west servers
	var isAlbionIP = false
	if strings.HasPrefix(state.GameServerIP, "5.188.125.") {
		// west server class c ip range
		serverID = 1
		isAlbionIP = true
	} else if strings.HasPrefix(state.GameServerIP, "5.45.187.") {
		// east server class c ip range
		isAlbionIP = true
		serverID = 2
	}

	// determine if the ConfigGlobal.PublicIngestBaseUrls contains either default east/west
	// data project server submission, if so, make sure it's set to the right hostname
	var westUrl = "http+pow://pow.west.albion-online-data.com"
	var eastUrl = "http+pow://pow.east.albion-online-data.com"
	if serverID == 1 && strings.Contains(ConfigGlobal.PublicIngestBaseUrls, eastUrl) {
		// we're on west but using east hostname, change it
		ConfigGlobal.PublicIngestBaseUrls = strings.ReplaceAll(ConfigGlobal.PublicIngestBaseUrls, eastUrl, westUrl)
	} else if serverID == 2 && strings.Contains(ConfigGlobal.PublicIngestBaseUrls, westUrl) {
		// we're on east but using west hostname, change it
		ConfigGlobal.PublicIngestBaseUrls = strings.ReplaceAll(ConfigGlobal.PublicIngestBaseUrls, westUrl, eastUrl)
	}

	// if this was a known albion online server ip, then let's log it
	if isAlbionIP {
		log.Tracef("Using %v for PublicIngestBaseUrls", ConfigGlobal.PublicIngestBaseUrls)
		log.Tracef("Returning Server ID %v (ip src: %v)", serverID, state.GameServerIP)
	}

	return serverID
}
