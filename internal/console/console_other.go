// internal/console/console_other.go
//go:build !windows

package console

// Supported reports whether console hiding is implemented on this
// platform. It is false everywhere except Windows: on macOS/Linux, a
// process launched from a terminal has no console window of its own to
// hide — the terminal is a separate, unrelated application.
func Supported() bool {
	return false
}

// Hide is a no-op on platforms without a distinct console window concept.
func Hide() {}

// Show is a no-op on platforms without a distinct console window concept.
func Show() {}

// Hidden always reports false on platforms without a distinct console
// window concept.
func Hidden() bool {
	return false
}

// Owned always reports false on platforms without a distinct console
// window concept - there's nothing to own.
func Owned() bool {
	return false
}
