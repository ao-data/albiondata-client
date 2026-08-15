package photon

import (
	"bytes"
	"encoding/binary"
)

const (
	photonHeaderLength   = 12
	commandHeaderLength  = 12
	fragmentHeaderLength = 20
)

// Photon command type constants
const (
	cmdDisconnect     = byte(4)
	cmdSendReliable   = byte(6)
	cmdSendUnreliable = byte(7)
	cmdSendFragment   = byte(8)
)

// Photon reliable message type constants (exported for tests)
const (
	MsgRequest  = byte(2)
	MsgResponse = byte(3)
	MsgEvent    = byte(4)
	// Some Albion builds send OperationResponse as type 7
	msgResponseAlt = byte(7)
	// Encrypted command (e.g. market data in some builds)
	msgEncrypted = byte(131)
)

// Photon packet header flag values.
const (
	flagEncrypted  = byte(1)
	flagCrcEnabled = byte(0xCC)
	crcFieldLength = 4
)

type segmentedPackage struct {
	totalLength  int
	bytesWritten int
	payload      []byte
}

// RawPacket holds the raw UDP/TCP payload of one Photon packet.
// Used for offline recording and replay.
type RawPacket struct {
	Payload []byte
}

// PhotonParser parses raw Photon UDP/TCP payloads and fires callbacks for each
// decoded operation or event. It is a port of PhotonParser.cs from
// https://github.com/JPCodeCraft/AlbionDataAvalonia.
type PhotonParser struct {
	pendingSegments map[int]*segmentedPackage

	// OnRequest is called for every decoded OperationRequest.
	OnRequest func(operationCode byte, params map[byte]interface{})

	// OnResponse is called for every decoded OperationResponse.
	OnResponse func(operationCode byte, returnCode int16, debugMessage string, params map[byte]interface{})

	// OnEvent is called for every decoded EventData.
	OnEvent func(code byte, params map[byte]interface{})

	// OnEncrypted is called when an encrypted message is encountered.
	OnEncrypted func()
}

// NewPhotonParser creates a PhotonParser wired to the given callbacks.
func NewPhotonParser(
	onRequest func(byte, map[byte]interface{}),
	onResponse func(byte, int16, string, map[byte]interface{}),
	onEvent func(byte, map[byte]interface{}),
) *PhotonParser {
	return &PhotonParser{
		pendingSegments: make(map[int]*segmentedPackage),
		OnRequest:       onRequest,
		OnResponse:      onResponse,
		OnEvent:         onEvent,
	}
}

// ReceivePacket processes a raw Photon UDP/TCP payload and fires the appropriate
// callbacks. Returns true if every Photon packet found in the payload was valid.
//
// A single UDP datagram can carry more than one Photon packet back-to-back
// (e.g. due to NIC/driver receive coalescing on some users' machines), so this
// scans for packet boundaries using the peerId/challenge identity shared by
// every packet in the payload and processes each one in turn.
func (p *PhotonParser) ReceivePacket(payload []byte) bool {
	if len(payload) < photonHeaderLength {
		return false
	}

	peerID, challenge, ok := getPacketIdentity(payload, 0)
	if !ok {
		return false
	}

	firstLen, isTerminalEncrypted, ok := getPacketLength(payload, 0)
	if !ok {
		return false
	}

	if isTerminalEncrypted || firstLen == len(payload) {
		return p.receiveSinglePacket(payload)
	}

	type packetRange struct{ offset, length int }
	ranges := []packetRange{{0, firstLen}}
	off := firstLen
	framingOK := true

	for off < len(payload) {
		if !matchesIdentity(payload, off, peerID, challenge) {
			framingOK = false
			break
		}
		length, isTerminal, valid := getPacketLength(payload, off)
		if !valid {
			framingOK = false
			break
		}
		ranges = append(ranges, packetRange{off, length})
		off += length
		if isTerminal {
			break
		}
	}

	success := framingOK
	for _, r := range ranges {
		if !p.receiveSinglePacket(payload[r.offset : r.offset+r.length]) {
			success = false
		}
	}
	return success
}

// receiveSinglePacket parses exactly one Photon packet (header + commands).
func (p *PhotonParser) receiveSinglePacket(payload []byte) bool {
	if len(payload) < photonHeaderLength {
		return false
	}

	offset := 2 // skip peerId (2 bytes)
	flags := payload[offset]
	offset++
	commandCount := int(payload[offset])
	offset++
	offset += 8 // skip timestamp (4) + challenge (4)

	if flags == flagEncrypted {
		if p.OnEncrypted != nil {
			p.OnEncrypted()
		}
		return false
	}

	if flags == flagCrcEnabled {
		if !available(payload, offset, crcFieldLength) {
			return false
		}
		crc := binary.BigEndian.Uint32(payload[offset:])
		offset += crcFieldLength

		crcPayload := append([]byte(nil), payload...)
		for i := photonHeaderLength; i < photonHeaderLength+crcFieldLength; i++ {
			crcPayload[i] = 0
		}
		if crc != calculateCrc(crcPayload) {
			return false
		}
	}

	for i := 0; i < commandCount; i++ {
		newOffset, ok := p.handleCommand(payload, offset)
		if !ok {
			return false
		}
		offset = newOffset
	}

	return offset == len(payload)
}

// getPacketIdentity reads the peerId/challenge pair that is common to every
// Photon packet coalesced into the same UDP payload.
func getPacketIdentity(payload []byte, offset int) (peerID int16, challenge int32, ok bool) {
	if !available(payload, offset, photonHeaderLength) {
		return 0, 0, false
	}
	peerID = int16(binary.BigEndian.Uint16(payload[offset:]))
	challenge = int32(binary.BigEndian.Uint32(payload[offset+8:]))
	return peerID, challenge, true
}

func matchesIdentity(payload []byte, offset int, expectedPeerID int16, expectedChallenge int32) bool {
	peerID, challenge, ok := getPacketIdentity(payload, offset)
	return ok && peerID == expectedPeerID && challenge == expectedChallenge
}

// getPacketLength scans the packet starting at offset (header + its declared
// commands) to determine where it ends, without decoding command contents.
// isTerminalEncrypted reports an encrypted packet, which has no discoverable
// length beyond "the rest of the payload" and must be the last packet.
func getPacketLength(payload []byte, offset int) (length int, isTerminalEncrypted bool, ok bool) {
	if !available(payload, offset, photonHeaderLength) {
		return 0, false, false
	}

	flags := payload[offset+2]
	if flags == flagEncrypted {
		return len(payload) - offset, true, true
	}

	headerLen := photonHeaderLength
	if flags == flagCrcEnabled {
		headerLen += crcFieldLength
	}
	if !available(payload, offset, headerLen) {
		return 0, false, false
	}

	commandCount := int(payload[offset+3])
	if commandCount == 0 {
		return 0, false, false
	}

	cmdOffset := offset + headerLen
	for i := 0; i < commandCount; i++ {
		if !available(payload, cmdOffset, commandHeaderLength) {
			return 0, false, false
		}
		cmdLen := int(binary.BigEndian.Uint32(payload[cmdOffset+4:]))
		if cmdLen < commandHeaderLength || !available(payload, cmdOffset, cmdLen) {
			return 0, false, false
		}
		cmdOffset += cmdLen
	}

	length = cmdOffset - offset
	if length < headerLen {
		return 0, false, false
	}
	return length, false, true
}

// calculateCrc reproduces Photon's CRC32 variant (standard IEEE polynomial,
// no final XOR) used to validate CRC-enabled packets (flags == 0xCC).
func calculateCrc(data []byte) uint32 {
	const key = uint32(3988292384)
	result := uint32(0xFFFFFFFF)
	for _, b := range data {
		result ^= uint32(b)
		for j := 0; j < 8; j++ {
			if result&1 != 0 {
				result = result>>1 ^ key
			} else {
				result >>= 1
			}
		}
	}
	return result
}

func (p *PhotonParser) handleCommand(src []byte, offset int) (int, bool) {
	if !available(src, offset, commandHeaderLength) {
		return offset, false
	}

	cmdType := src[offset]
	offset++
	offset++ // channelId
	offset++ // commandFlags
	offset++ // reserved byte
	cmdLen := int(binary.BigEndian.Uint32(src[offset:]))
	offset += 4
	offset += 4 // reliableSequenceNumber
	cmdLen -= commandHeaderLength

	if cmdLen < 0 || !available(src, offset, cmdLen) {
		return offset, false
	}

	switch cmdType {
	case cmdDisconnect:
		return offset + cmdLen, true

	case cmdSendUnreliable:
		if cmdLen < 4 {
			return offset + cmdLen, false
		}
		offset += 4
		cmdLen -= 4
		newOffset, _ := p.handleSendReliable(src, offset, cmdLen)
		return newOffset, true

	case cmdSendReliable:
		newOffset, _ := p.handleSendReliable(src, offset, cmdLen)
		return newOffset, true

	case cmdSendFragment:
		return p.handleSendFragment(src, offset, cmdLen), true

	default:
		return offset + cmdLen, true
	}
}

func (p *PhotonParser) handleSendReliable(src []byte, offset, cmdLen int) (int, bool) {
	if cmdLen < 2 || !available(src, offset, cmdLen) {
		return offset + cmdLen, false
	}

	/* signalByte := src[offset] */
	offset++
	msgType := src[offset]
	offset++
	cmdLen -= 2

	if !available(src, offset, cmdLen) {
		return offset + cmdLen, false
	}

	if msgType == msgEncrypted {
		if p.OnEncrypted != nil {
			p.OnEncrypted()
		}
		return offset + cmdLen, true
	}

	data := src[offset : offset+cmdLen]
	offset += cmdLen

	switch msgType {
	case MsgRequest:
		p.dispatchRequest(data)
	case MsgResponse, msgResponseAlt:
		p.dispatchResponse(data)
	case MsgEvent:
		p.dispatchEvent(data)
	}

	return offset, true
}

func (p *PhotonParser) dispatchRequest(data []byte) {
	if len(data) < 1 {
		return
	}
	opCode := data[0]
	params := deserializeParameterTable(data[1:])
	if p.OnRequest != nil {
		p.OnRequest(opCode, params)
	}
}

func (p *PhotonParser) dispatchResponse(data []byte) {
	if len(data) < 3 {
		return
	}
	opCode := data[0]
	returnCode := int16(binary.LittleEndian.Uint16(data[1:3]))

	// Per Protocol18Deserializer.DeserializeOperationResponse: after returnCode,
	// if bytes remain read one type byte + its value (the debug message), then
	// the parameter table.
	buf := bytes.NewBuffer(data[3:])
	debugMsg := ""

	if buf.Len() > 0 {
		tc, _ := buf.ReadByte()
		val := deserialize(buf, tc)
		switch v := val.(type) {
		case string:
			debugMsg = v
		case []string:
			// Albion embeds market-order data as a typed string array in the
			// position where the debug message would normally be.
			// Surface it as params[0] so the listener can identify and route it.
			params := map[byte]interface{}{0: v}
			if p.OnResponse != nil {
				p.OnResponse(opCode, returnCode, "", params)
			}
			return
		}
	}

	params := readParameterTable(buf)
	if p.OnResponse != nil {
		p.OnResponse(opCode, returnCode, debugMsg, params)
	}
}

func (p *PhotonParser) dispatchEvent(data []byte) {
	if len(data) < 1 {
		return
	}
	code := data[0]
	params := deserializeParameterTable(data[1:])
	if p.OnEvent != nil {
		p.OnEvent(code, params)
	}
}

func (p *PhotonParser) handleSendFragment(src []byte, offset, cmdLen int) int {
	if cmdLen < fragmentHeaderLength || !available(src, offset, fragmentHeaderLength) {
		return offset + cmdLen
	}

	startSeq := int(binary.BigEndian.Uint32(src[offset:]))
	offset += 4
	cmdLen -= 4
	offset += 4 // fragmentCount
	cmdLen -= 4
	offset += 4 // fragmentNumber
	cmdLen -= 4
	totalLen := int(binary.BigEndian.Uint32(src[offset:]))
	offset += 4
	cmdLen -= 4
	fragOffset := int(binary.BigEndian.Uint32(src[offset:]))
	offset += 4
	cmdLen -= 4

	fragLen := cmdLen
	if fragLen < 0 || !available(src, offset, fragLen) {
		return offset + fragLen
	}

	seg, ok := p.pendingSegments[startSeq]
	if !ok {
		seg = &segmentedPackage{
			totalLength: totalLen,
			payload:     make([]byte, totalLen),
		}
		p.pendingSegments[startSeq] = seg
	}

	end := fragOffset + fragLen
	if end <= len(seg.payload) {
		copy(seg.payload[fragOffset:end], src[offset:offset+fragLen])
	}
	offset += fragLen
	seg.bytesWritten += fragLen

	if seg.bytesWritten >= seg.totalLength {
		delete(p.pendingSegments, startSeq)
		p.handleSendReliable(seg.payload, 0, len(seg.payload))
	}

	return offset
}

// available reports whether src[offset:offset+count] is in bounds.
func available(src []byte, offset, count int) bool {
	return count >= 0 && offset >= 0 && len(src)-offset >= count
}
