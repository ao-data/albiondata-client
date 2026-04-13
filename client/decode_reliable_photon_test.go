package client

import (
	"testing"

	photon "github.com/ao-data/photon-spectator"
)

// When compact parsing is chosen because 253 decodes only as a varint there,
// remaining parameters must still use Photon type bytes (e.g. StringType=115),
// not be mis-read as inline int16.
func TestDecodeReliableMessage_compactPathDecodesStringParams(t *testing.T) {
	// String length is uint16 little-endian per Photon Protocol18 / StreamBuffer.
	data := []byte{
		253, 0x02, // operation code 2 (opJoin) as single-byte compact uint
		8, photon.StringType, 2, 0, 'h', 'i',
	}
	msg := photon.ReliableMessage{
		Type:           photon.OperationResponse,
		ParameterCount: 2,
		Data:           data,
	}
	params := photon.DecodeReliableMessage(msg)
	code, ok := toUint16(params[253])
	if !ok || code != 2 {
		t.Fatalf("params[253]=%v (%T), want uint16-compatible 2 (opJoin)", params[253], params[253])
	}
	s, ok := params[8].(string)
	if !ok || s != "hi" {
		t.Fatalf("params[8]=%v (%T), want string hi", params[8], params[8])
	}
}
