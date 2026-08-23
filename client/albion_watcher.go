package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// rescanInterval is how often the watcher re-enumerates physical network
// interfaces to pick up any that appeared after startup - e.g. a VPN's
// virtual adapter created when the VPN is turned on after this process
// has already started, which otherwise would never be listened on until
// the app is restarted.
const rescanInterval = 10 * time.Second

type albionProcessWatcher struct {
	known     []int
	devices   []string
	listeners map[int][]*listener
	quit      chan bool
	r         *Router
}

func newAlbionProcessWatcher() *albionProcessWatcher {
	return &albionProcessWatcher{
		listeners: make(map[int][]*listener),
		quit:      make(chan bool),
		r:         newRouter(),
	}
}

func (apw *albionProcessWatcher) run() error {
	log.Print("Watching Albion")
	physicalInterfaces, err := getAllPhysicalInterface()
	if err != nil {
		return err
	}
	apw.devices = physicalInterfaces
	log.Infof("Will listen to these devices: %v", apw.devices)
	go apw.r.run()

	lastRescan := time.Now()
	for {
		select {
		case <-apw.quit:
			apw.closeWatcher()
			return nil
		default:
			if len(apw.listeners) == 0 {
				apw.createListeners()
			} else if time.Since(lastRescan) >= rescanInterval {
				apw.rescanForNewDevices()
				lastRescan = time.Now()
			}
			time.Sleep(time.Second)
		}
	}
}

func (apw *albionProcessWatcher) closeWatcher() {
	log.Print("Albion watcher closed")

	for port := range apw.listeners {
		for _, l := range apw.listeners[port] {
			l.stop()
		}

		delete(apw.listeners, port)
	}

	apw.r.quit <- true
}

// rescanForNewDevices re-enumerates physical network interfaces and
// starts a listener on any that weren't already known, without
// disturbing existing listeners - see rescanInterval's doc comment.
func (apw *albionProcessWatcher) rescanForNewDevices() {
	current, err := getAllPhysicalInterface()
	if err != nil {
		log.Errorf("Rescan for new network interfaces failed: %v", err)
		return
	}

	known := make(map[string]bool, len(apw.devices))
	for _, d := range apw.devices {
		known[d] = true
	}

	var newDevices []string
	for _, d := range current {
		if !known[d] {
			newDevices = append(newDevices, d)
		}
	}
	if len(newDevices) == 0 {
		return
	}

	log.Infof("Found new network interfaces, starting capture on them: %v", newDevices)
	apw.devices = append(apw.devices, newDevices...)

	for port := range apw.listeners {
		for _, device := range newDevices {
			l := newListener(apw.r)
			go l.startOnline(device, port)

			apw.listeners[port] = append(apw.listeners[port], l)
		}
	}
}

func (apw *albionProcessWatcher) createListeners() {
	filtered := [1]int{5056} // keep overdesign to listen on many ports

	for _, port := range filtered {
		for _, device := range apw.devices {
			l := newListener(apw.r)
			go l.startOnline(device, port)

			apw.listeners[port] = append(apw.listeners[port], l)
		}
	}
}
