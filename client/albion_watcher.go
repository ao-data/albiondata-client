package client

import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)

// rescanInterval is how often the watcher re-enumerates physical network
// interfaces to pick up any that appeared after startup - e.g. a VPN's
// virtual adapter created when the VPN is turned on after this process
// has already started, which otherwise would never be listened on until
// the app is restarted - and to stop listening on any that disappeared,
// e.g. that same adapter when the VPN is turned back off.
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
				apw.rescan()
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

// diffDevices compares the previously-known device list against a fresh
// enumeration, pulled out of rescan so it's unit-testable without real
// network interfaces (same split as pcapdriver's Check/decide).
func diffDevices(known, current []string) (newDevices, vanishedDevices []string) {
	currentSet := make(map[string]bool, len(current))
	for _, d := range current {
		currentSet[d] = true
	}
	knownSet := make(map[string]bool, len(known))
	for _, d := range known {
		knownSet[d] = true
	}

	for _, d := range current {
		if !knownSet[d] {
			newDevices = append(newDevices, d)
		}
	}
	for _, d := range known {
		if !currentSet[d] {
			vanishedDevices = append(vanishedDevices, d)
		}
	}
	return newDevices, vanishedDevices
}

// rescan re-enumerates physical network interfaces and applies the
// result - see rescanInterval's doc comment.
func (apw *albionProcessWatcher) rescan() {
	current, err := getAllPhysicalInterface()
	if err != nil {
		log.Errorf("Rescan for network interfaces failed: %v", err)
		return
	}
	apw.applyDeviceDiff(current)
}

// applyDeviceDiff starts a listener on any device in current that wasn't
// already known, and stops + drops the listener(s) for any known device
// that's no longer in current. Split out from rescan so it's testable
// without a real getAllPhysicalInterface() call.
func (apw *albionProcessWatcher) applyDeviceDiff(current []string) {
	newDevices, vanishedDevices := diffDevices(apw.devices, current)

	if len(newDevices) > 0 {
		log.Infof("Found new network interfaces, starting capture on them: %v", newDevices)
		apw.devices = append(apw.devices, newDevices...)

		for port := range apw.listeners {
			for _, device := range newDevices {
				l := newListener(apw.r)
				l.device = device
				go l.startOnline(device, port)

				apw.listeners[port] = append(apw.listeners[port], l)
			}
		}
	}

	if len(vanishedDevices) > 0 {
		log.Infof("Network interfaces disappeared, stopping capture on them: %v", vanishedDevices)
		vanishedSet := make(map[string]bool, len(vanishedDevices))
		for _, d := range vanishedDevices {
			vanishedSet[d] = true
		}

		for port, listeners := range apw.listeners {
			kept := listeners[:0]
			for _, l := range listeners {
				if vanishedSet[l.device] {
					l.stop()
				} else {
					kept = append(kept, l)
				}
			}
			apw.listeners[port] = kept
		}

		remaining := apw.devices[:0]
		for _, d := range apw.devices {
			if !vanishedSet[d] {
				remaining = append(remaining, d)
			}
		}
		apw.devices = remaining
	}
}

func (apw *albionProcessWatcher) createListeners() {
	filtered := [1]int{5056} // keep overdesign to listen on many ports

	for _, port := range filtered {
		for _, device := range apw.devices {
			l := newListener(apw.r)
			l.device = device
			go l.startOnline(device, port)

			apw.listeners[port] = append(apw.listeners[port], l)
		}
	}
}
