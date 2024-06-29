//go:build linux || darwin
// +build linux darwin

package client

import (
	"log"
	"net"
	"strings"
)

// Check if the interface is detected
func isInterfacePresent(_interface string) bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range interfaces {
		if iface.Name == _interface {
			return true
		}
	}
	return false
}

func parseGivenInterfaces(interfaces string) []string {
	// Split the input string by comma
	outInterfaces := strings.Split(interfaces, ",")
	if outInterfaces == nil {
		log.Fatal("Interfaces with name: %v not found, when parsed: %v", interfaces, outInterfaces)
	}

	return outInterfaces
}

// Gets all physical interfaces based on filter results, ignoring all VM, Loopback and Tunnel interfaces.
func getAllPhysicalInterface() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("Failed to get interfaces: %v", err)
	}

	var outInterfaces []string

	for _, _interface := range interfaces {
		if _interface.Flags&net.FlagLoopback == 0 && _interface.Flags&net.FlagUp == 1 && isPhysicalInterface(_interface.HardwareAddr.String()) {
			outInterfaces = append(outInterfaces, _interface.Name)
		}
	}

	return outInterfaces
}
