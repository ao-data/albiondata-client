package client

import (
	"regexp"
	"strings"
	"time"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/notification"
)

// CacheSize limit size of messages in cache
const CacheSize = 8192

type marketHistoryInfo struct {
	albionId  int32
	timescale lib.Timescale
	quality   uint8
}

// marketDataEncryptionCorrelationWindow bounds how long after a market data
// request an OnEncrypted packet is assumed to be that request's (now
// encrypted) response, rather than an unrelated encrypted packet on the
// same connection.
const marketDataEncryptionCorrelationWindow = 3 * time.Second

type albionState struct {
	LocationId                   string
	LocationString               string
	CharacterId                  lib.CharacterID
	CharacterName                string
	GameServerIP                 string
	AODataServerID               int
	AODataIngestBaseURL          string
	LastMarketDataRequestAt      time.Time
	MarketDataEncryptionNotified bool
	BanditEventLastTimeSubmitted time.Time
	FestivitiesLastTimeSubmitted time.Time

	// A lot of information is sent out but not contained in the response when requesting marketHistory (e.g. ID)
	// This information is stored in marketHistoryInfo
	// This array acts as a type of cache for that info
	// The index is the message number (param255) % CacheSize
	marketHistoryIDLookup [CacheSize]marketHistoryInfo
	// TODO could this be improved?!
}

func (state albionState) IsValidLocation() bool {
	var onlydigits = regexp.MustCompile(`^[0-9]+$`)

	switch {
	case state.LocationId == "":
		log.Error("The players location has not yet been set. Please transition zones so the location can be identified.")
		if !ConfigGlobal.Debug {
			notification.Push("The players location has not yet been set. Please transition zones so the location can be identified.")
		}
		return false

	case onlydigits.MatchString(state.LocationId):
		return true
	case strings.HasPrefix(state.LocationId, "BLACKBANK-"):
		return true
	case strings.HasSuffix(state.LocationId, "-HellDen"):
		return true
	case strings.HasSuffix(state.LocationId, "-Auction2"):
		return true
	default:
		log.Error("The players location is not valid. Please transition zones so the location can be fixed.")
		if !ConfigGlobal.Debug {
			notification.Push("The players location is not valid. Please transition zones so the location can be fixed.")
		}
		return false
	}
}

func (state albionState) GetServer() (int, string) {
	// default to 0
	var serverID = 0
	var AODataIngestBaseURL = ""

	// if we happen to have a server id stored in state, lets re-default to that
	if state.AODataServerID != 0 {
		serverID = state.AODataServerID
	}
	if state.AODataIngestBaseURL != "" {
		AODataIngestBaseURL = state.AODataIngestBaseURL
	}

	// we get packets from other than game servers, so determine if it's a game server
	// based on soruce ip and if its east/west servers
	var isAlbionIP = false
	if strings.HasPrefix(state.GameServerIP, "5.188.125.") {
		// west server class c ip range
		serverID = 1
		isAlbionIP = true
		AODataIngestBaseURL = "https+pow://pow.west.albion-online-data.com"
	} else if strings.HasPrefix(state.GameServerIP, "5.45.187.") {
		// east server class c ip range
		isAlbionIP = true
		serverID = 2
		AODataIngestBaseURL = "https+pow://pow.east.albion-online-data.com"
	} else if strings.HasPrefix(state.GameServerIP, "193.169.238.") {
		// eu server class c ip range
		isAlbionIP = true
		serverID = 3
		AODataIngestBaseURL = "https+pow://pow.europe.albion-online-data.com"
	}

	// if this was a known albion online server ip, then let's log it
	if isAlbionIP {
		log.Tracef("Returning Server ID %v (ip src: %v)", serverID, state.GameServerIP)
		log.Tracef("Returning AODataIngestBaseURL %v (ip src: %v)", AODataIngestBaseURL, state.GameServerIP)
	}

	return serverID, AODataIngestBaseURL
}

// RecordMarketDataRequest notes that a market data request was just sent,
// and re-arms the encryption notification guard so a new request can
// trigger a new notification even if a prior one already fired.
func (state *albionState) RecordMarketDataRequest(now time.Time) {
	state.LastMarketDataRequestAt = now
	state.MarketDataEncryptionNotified = false
}

// ShouldNotifyMarketDataEncrypted reports whether an OnEncrypted packet
// seen at now should be treated as this client's market data response
// coming back encrypted - i.e. it followed a market data request closely
// enough to be that request's response, and hasn't already been notified
// on. Calling this marks the guard so a given request only ever notifies
// once, even if further encrypted packets arrive within the same window.
func (state *albionState) ShouldNotifyMarketDataEncrypted(now time.Time) bool {
	if state.LastMarketDataRequestAt.IsZero() || state.MarketDataEncryptionNotified {
		return false
	}
	if now.Sub(state.LastMarketDataRequestAt) > marketDataEncryptionCorrelationWindow {
		return false
	}
	state.MarketDataEncryptionNotified = true
	return true
}
