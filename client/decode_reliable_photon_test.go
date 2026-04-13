package client

import (
	"encoding/binary"
	"testing"

	"github.com/ao-data/albiondata-client/client/photon"
)

// buildPhotonPacket wraps a single reliable-message payload into a minimal
// valid Photon UDP packet (12-byte header + 12-byte command header + data).
func buildPhotonPacket(msgType byte, opCode byte, payload []byte) []byte {
	// Photon packet header (12 bytes): peerId(2) flags(1) cmdCount(1) timestamp(4) challenge(4)
	hdr := []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}

	// Command data: signalByte(1) + msgType(1) + opCode(1) + payload
	cmdData := append([]byte{0x00, msgType, opCode}, payload...)

	// Command header (12 bytes): type(1)=6 chId(1) flags(1) reserved(1) length(4) seqNum(4)
	cmdHdr := make([]byte, 12)
	cmdHdr[0] = 6 // SendReliable
	binary.BigEndian.PutUint32(cmdHdr[4:], uint32(12+len(cmdData)))

	pkt := append(hdr, cmdHdr...)
	pkt = append(pkt, cmdData...)
	return pkt
}

// writeVarint writes a Protocol18 compressed-varint count byte.
func writeVarint(v uint32) []byte {
	if v < 0x80 {
		return []byte{byte(v)}
	}
	return []byte{byte(v&0x7F | 0x80), byte(v >> 7)}
}

// TestParser_OperationRequest_compactPath verifies that params[253] is decoded
// as a compact varint when present, and that subsequent parameters still use
// normal Protocol18 type bytes.
func TestParser_OperationRequest_compactPath(t *testing.T) {
	// Parameter table: count=2
	// key=253 type=0x1E(IntZero)  value=(none - zero sentinel)
	// key=8   type=0x73(StringType legacy? no) type=7(Protocol18 String) len=2 "hi"
	//
	// Using Protocol18 type 7 (String) with varint length 2.
	paramTable := []byte{}
	paramTable = append(paramTable, writeVarint(2)...) // count=2
	paramTable = append(paramTable, 253, photon.TypeIntZero)
	paramTable = append(paramTable, 8, photon.TypeString, 2, 'h', 'i')

	pkt := buildPhotonPacket(photon.MsgRequest, 2, paramTable)

	var gotParams map[byte]interface{}
	parser := photon.NewPhotonParser(
		func(_ byte, params map[byte]interface{}) { gotParams = params },
		nil,
		nil,
	)
	if !parser.ReceivePacket(pkt) {
		t.Fatal("ReceivePacket returned false")
	}

	code, ok := toUint16(gotParams[253])
	if !ok || code != 0 {
		t.Fatalf("params[253]=%v (%T), want int32(0) (IntZero)", gotParams[253], gotParams[253])
	}
	s, ok := gotParams[8].(string)
	if !ok || s != "hi" {
		t.Fatalf("params[8]=%v (%T), want string \"hi\"", gotParams[8], gotParams[8])
	}
}

// TestParser_OperationResponse_returnsReturnCode verifies that the return code
// is decoded little-endian and passed to the OnResponse callback.
func TestParser_OperationResponse_returnsReturnCode(t *testing.T) {
	// Response body after opCode: returnCode(2 LE) + NullType(debug) + empty param table
	body := []byte{
		0x05, 0x00, // returnCode = 5 (LE)
		photon.TypeNull,  // debug type = null (no debug string)
	}
	body = append(body, writeVarint(0)...) // param count = 0

	pkt := buildPhotonPacket(photon.MsgResponse, 42, body)

	var gotCode int16
	parser := photon.NewPhotonParser(
		nil,
		func(_ byte, rc int16, _ string, _ map[byte]interface{}) { gotCode = rc },
		nil,
	)
	parser.ReceivePacket(pkt)

	if gotCode != 5 {
		t.Fatalf("returnCode=%d, want 5", gotCode)
	}
}
