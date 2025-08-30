//go:build darwin
// +build darwin

package client

import (
	"net"
	"strings"

	"github.com/ao-data/albiondata-client/log"
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

func parseGivenInterfaces(interfaces string) ([]string, error) {
	// Split the input string by comma
	outInterfaces := strings.Split(interfaces, ",")

	for i, _interface := range outInterfaces {
		if !isInterfacePresent(_interface) {
			log.Debugf("Interface with name: %v not found, removing from list ...", _interface)
			// remove the interface from the list
			outInterfaces = append(outInterfaces[:i], outInterfaces[i+1:]...)
		}
	}

	return outInterfaces, nil
}

// Gets all physical interfaces based on filter results, ignoring all VM, Loopback and Tunnel interfaces.
func getAllPhysicalInterface() ([]string, error) {
	if ConfigGlobal.ListenDevices != "" {
		return parseGivenInterfaces(ConfigGlobal.ListenDevices)
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("Failed to get interfaces: %v", err)
	}

	var outInterfaces []string

	for _, _interface := range interfaces {
		if _interface.Flags&net.FlagLoopback == 0 && _interface.Flags&net.FlagUp == 1 && isPhysicalInterface(_interface.HardwareAddr.String()) {
			outInterfaces = append(outInterfaces, _interface.Name)
		}
	}

	return outInterfaces, nil
}
