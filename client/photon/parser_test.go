package photon

import (
	"encoding/binary"
	"testing"
)

// buildFragmentPacket wraps a single SendFragment command into a minimal
// valid Photon UDP packet (12-byte header + 12-byte command header +
// 20-byte fragment header + data).
func buildFragmentPacket(startSeq, fragCount, fragNum, totalLen, fragOffset uint32, data []byte) []byte {
	hdr := []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}

	fragHdr := make([]byte, fragmentHeaderLength)
	binary.BigEndian.PutUint32(fragHdr[0:], startSeq)
	binary.BigEndian.PutUint32(fragHdr[4:], fragCount)
	binary.BigEndian.PutUint32(fragHdr[8:], fragNum)
	binary.BigEndian.PutUint32(fragHdr[12:], totalLen)
	binary.BigEndian.PutUint32(fragHdr[16:], fragOffset)

	cmdData := append(fragHdr, data...)

	cmdHdr := make([]byte, commandHeaderLength)
	cmdHdr[0] = cmdSendFragment
	binary.BigEndian.PutUint32(cmdHdr[4:], uint32(commandHeaderLength+len(cmdData)))

	pkt := append(append([]byte{}, hdr...), cmdHdr...)
	pkt = append(pkt, cmdData...)
	return pkt
}

// TestPendingSegments_BoundedByCap verifies that incomplete fragment
// reassembly (e.g. because a fragment was lost to packet loss during a
// network hiccup) doesn't grow pendingSegments without bound - each
// abandoned start-of-fragment allocates a totalLen-sized buffer that is
// never released otherwise.
func TestPendingSegments_BoundedByCap(t *testing.T) {
	parser := NewPhotonParser(nil, nil, nil)

	for i := 0; i < maxPendingSegments*4; i++ {
		// Send only the first fragment of a 2-fragment message, so it never
		// completes and stays in pendingSegments.
		pkt := buildFragmentPacket(uint32(i), 2, 0, 20, 0, make([]byte, 10))
		parser.ReceivePacket(pkt)
	}

	if len(parser.pendingSegments) > maxPendingSegments {
		t.Fatalf("pendingSegments grew to %d entries, want capped at %d", len(parser.pendingSegments), maxPendingSegments)
	}
}

// TestPendingSegments_RejectsOversizedTotalLength verifies a corrupt or
// hostile totalLen field (read straight off the wire) can't force a huge
// allocation.
func TestPendingSegments_RejectsOversizedTotalLength(t *testing.T) {
	parser := NewPhotonParser(nil, nil, nil)

	pkt := buildFragmentPacket(1, 2, 0, maxSegmentTotalLength+1, 0, make([]byte, 10))
	parser.ReceivePacket(pkt)

	if len(parser.pendingSegments) != 0 {
		t.Fatalf("pendingSegments has %d entries, want 0 for an oversized totalLen", len(parser.pendingSegments))
	}
}
