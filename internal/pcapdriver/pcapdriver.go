// Package pcapdriver detects the system's packet-capture driver state on
// Windows - Npcap present, only the abandoned WinPcap present, or
// neither - so the client can warn the user on startup instead of
// silently failing to capture. This exists because the project's own
// Windows installer used to bundle WinPcap (unmaintained since 2013,
// unable to open Wintun/WireGuard-style VPN adapters at all) and many
// existing installs still have it instead of Npcap.
package pcapdriver

// Warning describes a problem with the system's packet-capture driver
// that the user should act on. The zero Warning means nothing is wrong.
// HelpURL is empty when there's nothing useful to link to (e.g. the fix
// is "restart as Administrator", not a download).
type Warning struct {
	Message string
	HelpURL string
}
