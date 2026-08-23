//go:build darwin

// Package dockicon toggles the app's macOS Dock icon.
package dockicon

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

static void albionSetDockIconVisible(bool visible) {
	if (visible) {
		[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
	} else {
		[NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
	}
}
*/
import "C"

// SetVisible shows or hides the app's Dock icon. The app starts as an
// accessory (menu-bar-only, per Mac.ActivationPolicy where the dashboard
// window is created) so it doesn't clutter the Dock while just running
// in the background; switching to Regular while the dashboard window is
// shown gives it a normal Dock presence for as long as that's true.
func SetVisible(visible bool) {
	C.albionSetDockIconVisible(C.bool(visible))
}
