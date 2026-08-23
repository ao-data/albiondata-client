package client

import (
	"reflect"
	"testing"
)

func TestDiffDevices_NoChange(t *testing.T) {
	known := []string{"eth0", "wlan0"}
	current := []string{"eth0", "wlan0"}

	newDevices, vanishedDevices := diffDevices(known, current)

	if len(newDevices) != 0 {
		t.Fatalf("expected no new devices, got %v", newDevices)
	}
	if len(vanishedDevices) != 0 {
		t.Fatalf("expected no vanished devices, got %v", vanishedDevices)
	}
}

func TestDiffDevices_NewDeviceAppears(t *testing.T) {
	known := []string{"eth0"}
	current := []string{"eth0", "PIA-Wintun"}

	newDevices, vanishedDevices := diffDevices(known, current)

	if !reflect.DeepEqual(newDevices, []string{"PIA-Wintun"}) {
		t.Fatalf("expected [PIA-Wintun] as new, got %v", newDevices)
	}
	if len(vanishedDevices) != 0 {
		t.Fatalf("expected no vanished devices, got %v", vanishedDevices)
	}
}

func TestDiffDevices_DeviceVanishes(t *testing.T) {
	known := []string{"eth0", "PIA-Wintun"}
	current := []string{"eth0"}

	newDevices, vanishedDevices := diffDevices(known, current)

	if len(newDevices) != 0 {
		t.Fatalf("expected no new devices, got %v", newDevices)
	}
	if !reflect.DeepEqual(vanishedDevices, []string{"PIA-Wintun"}) {
		t.Fatalf("expected [PIA-Wintun] as vanished, got %v", vanishedDevices)
	}
}

func TestDiffDevices_AppearsAndVanishesTogether(t *testing.T) {
	known := []string{"eth0", "old-vpn"}
	current := []string{"eth0", "new-vpn"}

	newDevices, vanishedDevices := diffDevices(known, current)

	if !reflect.DeepEqual(newDevices, []string{"new-vpn"}) {
		t.Fatalf("expected [new-vpn] as new, got %v", newDevices)
	}
	if !reflect.DeepEqual(vanishedDevices, []string{"old-vpn"}) {
		t.Fatalf("expected [old-vpn] as vanished, got %v", vanishedDevices)
	}
}

func TestDiffDevices_EmptyKnown(t *testing.T) {
	newDevices, vanishedDevices := diffDevices(nil, []string{"eth0"})

	if !reflect.DeepEqual(newDevices, []string{"eth0"}) {
		t.Fatalf("expected [eth0] as new, got %v", newDevices)
	}
	if len(vanishedDevices) != 0 {
		t.Fatalf("expected no vanished devices, got %v", vanishedDevices)
	}
}

// TestApplyDeviceDiff_StopsListenerForVanishedDevice exercises the real
// applyDeviceDiff (not just diffDevices) - a vanished device's listener
// must be stopped and dropped from apw.listeners/apw.devices, while a
// listener for a still-present device is left alone.
func TestApplyDeviceDiff_StopsListenerForVanishedDevice(t *testing.T) {
	apw := newAlbionProcessWatcher()
	apw.devices = []string{"eth0", "vanishing-vpn"}

	stayingListener := newListener(apw.r)
	stayingListener.device = "eth0"
	vanishingListener := newListener(apw.r)
	vanishingListener.device = "vanishing-vpn"
	apw.listeners[5056] = []*listener{stayingListener, vanishingListener}

	apw.applyDeviceDiff([]string{"eth0"})

	if len(apw.listeners[5056]) != 1 || apw.listeners[5056][0] != stayingListener {
		t.Fatalf("expected only the staying listener to remain, got %v", apw.listeners[5056])
	}
	if !reflect.DeepEqual(apw.devices, []string{"eth0"}) {
		t.Fatalf("expected devices to be [eth0], got %v", apw.devices)
	}
}

// TestApplyDeviceDiff_StartsListenerForNewDevice mirrors the appear case:
// a new device gets a listener on every currently-tracked port, and is
// added to apw.devices.
func TestApplyDeviceDiff_StartsListenerForNewDevice(t *testing.T) {
	apw := newAlbionProcessWatcher()
	apw.devices = []string{"eth0"}

	existing := newListener(apw.r)
	existing.device = "eth0"
	apw.listeners[5056] = []*listener{existing}

	apw.applyDeviceDiff([]string{"eth0", "new-vpn"})

	if len(apw.listeners[5056]) != 2 {
		t.Fatalf("expected 2 listeners on port 5056, got %d", len(apw.listeners[5056]))
	}
	if !reflect.DeepEqual(apw.devices, []string{"eth0", "new-vpn"}) {
		t.Fatalf("expected devices to be [eth0 new-vpn], got %v", apw.devices)
	}
}
