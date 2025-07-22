//go:build darwin

package client

import (
	"net"
	"strings"
)

// Gets all physical interfaces based on filter results, ignoring all VM, Loopback and Tunnel interfaces.
func getAllPhysicalInterface() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var outInterfaces []string
	wantedDevices := strings.Split(ConfigGlobal.ListenDevices, ",")

	if ConfigGlobal.ListenDevices != "" && len(wantedDevices) > 0 {
		for _, wanted := range wantedDevices {
			for _, iface := range interfaces {
				if iface.Name == strings.TrimSpace(wanted) {
					outInterfaces = append(outInterfaces, iface.Name)
				}
			}
		}
	} else {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagLoopback == 0 &&
				iface.Flags&net.FlagUp != 0 &&
				isPhysicalInterface(iface.HardwareAddr.String()) {
				outInterfaces = append(outInterfaces, iface.Name)
			}
		}
	}

	return outInterfaces, nil
}
