package client

import (
	"encoding/hex"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
	"github.com/mitchellh/mapstructure"
)

func toUint16(v interface{}) (uint16, bool) {
	switch val := v.(type) {
	case uint16:
		return val, true
	case int16:
		return uint16(val), true
	case int8:
		return uint16(int16(val)), true
	case uint8:
		return uint16(val), true
	case int32:
		return uint16(val), true
	case uint32:
		return uint16(val), true
	case int64:
		return uint16(val), true
	case uint64:
		return uint16(val), true
	case string:
		n, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return 0, false
		}
		return uint16(n), true
	}
	return 0, false
}

func decodeRequest(params map[uint8]interface{}) (operation operation, err error) {
	code, ok := resolveOperationCode(params)
	if !ok {
		return nil, nil
	}

	switch OperationType(code) {
	case opGetGameServerByCluster:
		operation = &operationGetGameServerByCluster{}
	case opAuctionGetOffers:
		operation = &operationAuctionGetOffers{}
	case opAuctionGetItemAverageStats:
		operation = &operationAuctionGetItemAverageStats{}
	case opGetClusterMapInfo:
		operation = &operationGetClusterMapInfo{}
	// case opGoldMarketGetAverageInfo:
	case opGoldMarketGetAverageInfo:
		operation = &operationGoldMarketGetAverageInfo{}
	case opRealEstateGetAuctionData:
		operation = &operationRealEstateGetAuctionData{}
	case opRealEstateBidOnAuction:
		operation = &operationRealEstateBidOnAuction{}
	default:
		return nil, nil
	}

	err = decodeParams(params, operation)

	return operation, err
}

func decodeResponse(params map[uint8]interface{}) (operation operation, err error) {
	code, ok := resolveOperationCode(params)
	if !ok {
		return nil, nil
	}

	// log.Infof("decodeResponse: %v, params: %v", code, params)

	switch OperationType(code) {
	case opJoin:
		operation = &operationJoinResponse{}
	case opGetGameServerByCluster:
		operation = &operationGetGameServerByCluster{}
	case opAuctionGetOffers:
		operation = &operationAuctionGetOffersResponse{}
	case opAuctionGetRequests:
		operation = &operationAuctionGetRequestsResponse{}
	case opAuctionBuyOffer:
		operation = &operationAuctionGetRequestsResponse{}
	case opAuctionGetItemAverageStats:
		operation = &operationAuctionGetItemAverageStatsResponse{}
	case opGetMailInfos:
		operation = &operationGetMailInfosResponse{}
	case opReadMail:
		operation = &operationReadMail{}
	case opGetClusterMapInfo:
		operation = &operationGetClusterMapInfoResponse{}
	// case opGoldMarketGetAverageInfo:
	case opGoldMarketGetAverageInfo:
		operation = &operationGoldMarketGetAverageInfoResponse{}
	case opRealEstateGetAuctionData:
		operation = &operationRealEstateGetAuctionDataResponse{}
	case opRealEstateBidOnAuction:
		operation = &operationRealEstateBidOnAuctionResponse{}
	default:
		if looksLikeJoinResponse(params) {
			if _, ok := params[8].(string); !ok {
				if location, found := extractLocationLikeString(params); found {
					params[8] = location
				}
			}
			operation = &operationJoinResponse{}
			break
		}
		return nil, nil
	}

	err = decodeParams(params, operation)

	return operation, err
}

func decodeEvent(params map[uint8]interface{}) (event operation, err error) {
	eventType, ok := resolveEventCode(params)
	if !ok {
		log.Debugf("decodeEvent: unexpected type for params[252]: %T = %s", params[252], formatDebugValue(params[252], 0))
		return nil, nil
	}

	// log.Infof("decodeEvent: %v, params: %v", eventType, params)

	switch EventType(eventType) {
	// case evRespawn: //TODO: confirm this eventCode (old 77)
	// 	event = &eventPlayerOnlineStatus{}
	// case evCharacterStats: //TODO: confirm this eventCode (old 114)
	// 	event = &eventSkillData{}
	//case evRedZonePlayerNotification:
	//	event = &eventRedZonePlayerNotification{}
	case evRedZoneWorldMapEvent:
		event = &eventRedZoneWorldMapEvent{}
	default:
		return nil, nil
	}

	err = decodeParams(params, event)

	return event, err
}

func decodeParams(params map[uint8]interface{}, operation operation) error {
	convertGameObjects := func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
		if from == reflect.TypeOf([]int8{}) && to == reflect.TypeOf(lib.CharacterID("")) {
			log.Debug("Parsing character ID from mixed-endian UUID")

			return decodeCharacterID(v.([]int8)), nil
		}

		return v, nil
	}

	config := mapstructure.DecoderConfig{
		DecodeHook: convertGameObjects,
		Result:     operation,
	}

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return err
	}

	// Decided that the maps were easier to work with in most places with uint8 keys
	// Therefore we have to convert to a string map in order for the decode to work here
	// Should be negligible performance loss
	stringMap := make(map[string]interface{})
	for k, v := range params {
		stringMap[strconv.Itoa(int(k))] = v
	}

	err = decoder.Decode(stringMap)

	return err
}

func decodeCharacterID(array []int8) lib.CharacterID {
	/* So this is a UUID, which is stored in a 'mixed-endian' format.
	The first three components are stored in little-endian, the rest in big-endian.
	See https://en.wikipedia.org/wiki/Universally_unique_identifier#Encoding.
	By default, our int array is read as big-endian, so we need to swap the first
	three components of the UUID
	*/
	b := make([]byte, len(array))

	// First, convert to byte
	for k, v := range array {
		b[k] = byte(v)
	}

	// swap first component
	b[0], b[1], b[2], b[3] = b[3], b[2], b[1], b[0]

	// swap second component
	b[4], b[5] = b[5], b[4]

	// swap third component
	b[6], b[7] = b[7], b[6]

	// format it UUID-style
	var buf [36]byte
	hex.Encode(buf[:], b[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], b[10:])

	return lib.CharacterID(buf[:])
}

func normalizeOperationCode(code uint16) uint16 {
	if isKnownOperationCode(code) {
		return code
	}
	swapped := (code << 8) | (code >> 8)
	if isKnownOperationCode(swapped) {
		return swapped
	}
	// Common post-update artifact: values like 0xDA00 where real code is 0x00DA.
	if code > 0x00FF && code&0x00FF == 0 {
		return code >> 8
	}
	return code
}

func normalizeEventCode(code uint16) uint16 {
	if isKnownEventCode(code) {
		return code
	}
	swapped := (code << 8) | (code >> 8)
	if isKnownEventCode(swapped) {
		return swapped
	}
	if code > 0x00FF && code&0x00FF == 0 {
		return code >> 8
	}
	return code
}

func isKnownOperationCode(code uint16) bool {
	name := OperationType(code).String()
	return !strings.HasPrefix(name, "OperationType(")
}

func isKnownEventCode(code uint16) bool {
	name := EventType(code).String()
	return !strings.HasPrefix(name, "EventType(")
}

func looksLikeJoinResponse(params map[uint8]interface{}) bool {
	if location, ok := params[8].(string); ok {
		normalized := normalizeLocationID(location)
		if normalized != "" {
			return true
		}
	}
	if _, found := extractLocationLikeString(params); found {
		return true
	}
	for _, value := range params {
		s, ok := value.(string)
		if !ok {
			continue
		}
		ls := strings.ToLower(s)
		if strings.Contains(s, "@player-island") || strings.Contains(s, "@island-") || strings.Contains(ls, "island") {
			return true
		}
	}
	return false
}

func extractLocationLikeString(v interface{}) (string, bool) {
	switch value := v.(type) {
	case map[uint8]interface{}:
		for _, vv := range value {
			if s, ok := extractLocationLikeString(vv); ok {
				return s, true
			}
		}
	case map[string]interface{}:
		for _, vv := range value {
			if s, ok := extractLocationLikeString(vv); ok {
				return s, true
			}
		}
	case []interface{}:
		for _, vv := range value {
			if s, ok := extractLocationLikeString(vv); ok {
				return s, true
			}
		}
	case []byte:
		return extractLocationLikeFromText(string(value))
	case []int8:
		b := make([]byte, len(value))
		for i, n := range value {
			b[i] = byte(n)
		}
		return extractLocationLikeFromText(string(b))
	case []int:
		b := make([]byte, len(value))
		for i, n := range value {
			if n < 0 || n > 255 {
				return "", false
			}
			b[i] = byte(n)
		}
		return extractLocationLikeFromText(string(b))
	case string:
		return extractLocationLikeFromText(value)
	}
	return "", false
}

func extractLocationLikeFromText(text string) (string, bool) {
	if text == "" {
		return "", false
	}
	if normalized := normalizeLocationID(text); normalized != "" {
		return normalized, true
	}
	// Split on non-printable runes and scan printable chunks.
	var chunk strings.Builder
	flush := func() (string, bool) {
		if chunk.Len() == 0 {
			return "", false
		}
		s := chunk.String()
		chunk.Reset()
		if normalized := normalizeLocationID(s); normalized != "" && normalized != s {
			return normalized, true
		}
		ls := strings.ToLower(s)
		if strings.Contains(s, "@player-island") || strings.Contains(s, "@island-") || strings.Contains(ls, "island") {
			return normalizeLocationID(s), true
		}
		return "", false
	}
	for _, r := range text {
		if unicode.IsPrint(r) {
			chunk.WriteRune(r)
			continue
		}
		if s, ok := flush(); ok {
			return s, true
		}
	}
	return flush()
}

func resolveOperationCode(params map[uint8]interface{}) (uint16, bool) {
	v, ok := params[253]
	if !ok {
		return 0, false
	}
	code, ok := toUint16(v)
	if !ok {
		return 0, false
	}
	return normalizeOperationCode(code), true
}

func resolveEventCode(params map[uint8]interface{}) (uint16, bool) {
	v, ok := params[252]
	if !ok {
		return 0, false
	}
	code, ok := toUint16(v)
	if !ok {
		return 0, false
	}
	return normalizeEventCode(code), true
}
