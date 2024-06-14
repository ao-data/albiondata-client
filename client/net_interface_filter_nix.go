// +build linux

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
       // -l option was given, filter for explicit wanted devices
       if (ConfigGlobal.ListenDevices != "") && len(wantedDevices) > 0 {
               for _, wantedDevice := range wantedDevices {
                       for _, inter := range interfaces {
                               if inter.Name == wantedDevice {
                                       outInterfaces = append(outInterfaces, inter.Name)
                               }

                       }
               }
       // NO -l option was given, try to find all physical devices
       } else {
               for _, _interface := range interfaces {
                       if _interface.Flags&net.FlagLoopback == 0 && _interface.Flags&net.FlagUp == 1 && isPhysicalInterface(_interface.HardwareAddr.String()) {
                               outInterfaces = append(outInterfaces, _interface.Name)
                       }
               }
       }



       return outInterfaces, nil
}
