//go:build !windows

package pcapdriver

// Check always returns the zero Warning outside Windows - macOS and
// Linux use libpcap directly with no separate driver-install step to
// detect.
func Check() Warning { return Warning{} }
