// +build windows

package client

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	IfOperStatusUp            = 1
	IF_TYPE_SOFTWARE_LOOPBACK = 24
	IF_TYPE_TUNNEL            = 131
)

const hexDigit = "0123456789abcdef"

func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*windows.IpAdapterAddresses
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}

func bytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

// physicalAddrToString formats the first addrLen bytes of physAddr as a
// colon-separated hex string. addrLen must be respected rather than
// always formatting the full fixed-size array: Windows leaves
// PhysicalAddress zeroed for adapters with no real hardware address (its
// length is reported separately via PhysicalAddressLength) - such as
// Wintun, the Layer-3-only virtual adapter PIA's Windows client defaults
// to for its VPN tunnel. Formatting all 8 zero bytes unconditionally
// produces "00:00:00:00:00:00:00:00", which collides with the
// "00:00:00:00:00" Teredo-pseudo-interface entry in
// macAddrPartsToFilter and gets the tunnel wrongly excluded from
// capture - exactly the interface carrying all VPN'd game traffic.
func physicalAddrToString(physAddr [8]byte, addrLen uint32) string {
	if addrLen == 0 || addrLen > uint32(len(physAddr)) {
		return ""
	}
	addr := physAddr[:addrLen]
	buf := make([]byte, 0, len(addr)*3-1)
	for i, b := range addr {
		if i > 0 {
			buf = append(buf, ':')
		}
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}
	return string(buf)
}

func cStringToString(cs *uint16) (s string) {
	if cs != nil {
		us := make([]uint16, 0, 256)
		for p := uintptr(unsafe.Pointer(cs)); ; p += 2 {
			u := *(*uint16)(unsafe.Pointer(p))
			if u == 0 {
				return string(utf16.Decode(us))
			}
			us = append(us, u)
		}
	}
	return ""
}

// Gets all physical interfaces based on filter results, ignoring all VM, Loopback and Tunnel interfaces.
func getAllPhysicalInterface() ([]string, error) {
	aa, err := adapterAddresses()
	if err != nil {
		return nil, err
	}

	var outInterfaces []string
	devices := strings.Split(strings.ReplaceAll(strings.ToLower(ConfigGlobal.ListenDevices), "-", ":"), ",")

	for _, pa := range aa {
		mac := physicalAddrToString(pa.PhysicalAddress, pa.PhysicalAddressLength)
		deviceFound := false
		if len(devices) > 0 {
			for _, device := range devices {
				if strings.HasPrefix(strings.ToLower(mac), device) {
					deviceFound = true
					break
				}
			}
			if !deviceFound {
				continue
			}
		}
		name := "\\Device\\NPF_" + bytePtrToString(pa.AdapterName)

		if pa.IfType != uint32(IF_TYPE_SOFTWARE_LOOPBACK) && pa.IfType != uint32(IF_TYPE_TUNNEL) &&
			pa.OperStatus == uint32(IfOperStatusUp) && isPhysicalInterface(mac) {
			outInterfaces = append(outInterfaces, name)
		}
	}
	if len(outInterfaces) == 0 {
		if len(devices) > 0 {
			return nil, errors.New("mac address was not found")
		} else {
			return nil, errors.New("could not find a network interface")
		}
	}

	return outInterfaces, nil
}
