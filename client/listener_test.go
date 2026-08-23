package client

import "testing"

// A device that can't be opened for capture (e.g. a Wintun-based VPN
// adapter on Windows, or simply a nonexistent name) must not crash the
// whole process - startOnline runs on its own unrecovered goroutine (see
// albionProcessWatcher.createListeners), so a panic here previously took
// down capture on every other interface too.
func TestListener_StartOnline_UnopenableDeviceDoesNotPanic(t *testing.T) {
	l := newListener(newRouter())

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("startOnline panicked on an unopenable device: %v", r)
		}
	}()

	l.startOnline("this-device-does-not-exist", 5056)

	if l.handle != nil {
		t.Fatalf("expected handle to remain nil after a failed open, got %+v", l.handle)
	}
}

func TestListener_Stop_NilHandleDoesNotPanic(t *testing.T) {
	l := newListener(newRouter())

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("stop panicked on a listener whose handle was never opened: %v", r)
		}
	}()

	l.stop()
}
