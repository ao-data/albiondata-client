//go:build !darwin

// Package dockicon toggles the app's macOS Dock icon.
package dockicon

// SetVisible is a no-op outside macOS - Windows and Linux always show a
// taskbar entry for a visible window, with no separate concept of a
// background-only "accessory" app to toggle.
func SetVisible(visible bool) {}
