// +build windows

package client

import "testing"

func TestPhysicalAddrToString_ZeroLengthAddressIsEmpty(t *testing.T) {
	// Adapters with no real hardware address (e.g. Wintun, the
	// Layer-3-only virtual adapter PIA's Windows client defaults to)
	// report PhysicalAddressLength=0 with a zeroed PhysicalAddress
	// array. This must format as "", not as a fabricated all-zero MAC -
	// see physicalAddrToString's doc comment for why that previously
	// caused such adapters to be misidentified as Teredo
	// pseudo-interfaces and excluded from capture.
	var addr [8]byte
	got := physicalAddrToString(addr, 0)
	if got != "" {
		t.Fatalf("physicalAddrToString(zeroed, 0) = %q, want empty string", got)
	}
	if !isPhysicalInterface(got) {
		t.Fatalf("isPhysicalInterface(%q) = false, want true (must not be excluded as a VM/Teredo interface)", got)
	}
}

func TestPhysicalAddrToString_FormatsRealMacAddress(t *testing.T) {
	addr := [8]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	got := physicalAddrToString(addr, 6)
	want := "aa:bb:cc:dd:ee:ff"
	if got != want {
		t.Fatalf("physicalAddrToString(addr, 6) = %q, want %q", got, want)
	}
}

func TestPhysicalAddrToString_LengthOutOfRangeIsEmpty(t *testing.T) {
	var addr [8]byte
	got := physicalAddrToString(addr, 9)
	if got != "" {
		t.Fatalf("physicalAddrToString(addr, 9) = %q, want empty string for out-of-range length", got)
	}
}
