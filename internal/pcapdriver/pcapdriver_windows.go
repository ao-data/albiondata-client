//go:build windows

package pcapdriver

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const npcapDownloadURL = "https://npcap.com/#download"

// npcapGuidance is shared between the two "go install it" cases below.
// All-users mode is the recommended default: it lets this client capture
// without running elevated, at the cost that any other application on
// the system could also read network traffic without needing admin
// rights either - a tradeoff worth stating plainly rather than silently
// picking for the user.
const npcapGuidance = "Install Npcap to enable network capture. During setup, we " +
	"recommend leaving \"Restrict Npcap driver's access to Administrators only\" " +
	"unchecked - this client can then capture without running as Administrator, " +
	"though it also means any other application on this system could read network " +
	"traffic without admin rights. If you'd rather restrict that, check the box, " +
	"but you'll need to start this client as Administrator from then on."

// Check inspects the system's packet-capture driver and returns a
// Warning if the user should do something about it, or the zero Warning
// if everything looks fine.
func Check() Warning {
	npcapPresent, npcapAdminOnly := npcapState()
	elevated := windows.GetCurrentProcessToken().IsElevated()
	return decide(npcapPresent, npcapAdminOnly, winPcapPresent(), elevated)
}

// decide is Check's decision logic, pulled out from the registry/token
// reads above it so it can be unit tested without depending on the
// actual system's driver state.
func decide(npcapPresent, npcapAdminOnly, winPcapPresent, processElevated bool) Warning {
	if npcapPresent {
		if npcapAdminOnly && !processElevated {
			return Warning{
				Message: "Npcap is installed in admin-only mode, but this client isn't " +
					"running as Administrator. Network capture will not work until you " +
					"restart this client as Administrator, or reinstall Npcap with " +
					"\"Restrict Npcap driver's access to Administrators only\" unchecked.",
			}
		}
		return Warning{}
	}

	if winPcapPresent {
		return Warning{
			Message: "Only the outdated WinPcap driver was detected. This client's installer " +
				"used to install WinPcap, but we've moved to Npcap - WinPcap has been " +
				"unsupported since 2013 and cannot capture on modern VPN adapters " +
				"(Wintun/WireGuard). " + npcapGuidance,
			HelpURL: npcapDownloadURL,
		}
	}

	return Warning{
		Message: "No packet capture driver was detected. " + npcapGuidance,
		HelpURL: npcapDownloadURL,
	}
}

// npcapState reports whether Npcap is installed and, if so, whether it
// was installed restricted to Administrators only (its "AdminOnly"
// registry value). Checks the WOW6432Node fallback too since this
// process may be running 32-bit under WOW64 on 64-bit Windows, same as
// Npcap's own installer.
func npcapState() (present bool, adminOnly bool) {
	for _, path := range []string{`SOFTWARE\Npcap`, `SOFTWARE\WOW6432Node\Npcap`} {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, verr := k.GetIntegerValue("AdminOnly")
		k.Close()
		return true, verr == nil && v != 0
	}
	return false, false
}

// winPcapPresent reports whether the legacy WinPcap driver is installed,
// checked via its installer's own registry key rather than a
// version-specific detail, so it isn't tied to exactly 4.1.3.
func winPcapPresent() bool {
	for _, path := range []string{`SOFTWARE\WinPcap`, `SOFTWARE\WOW6432Node\WinPcap`} {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.QUERY_VALUE)
		if err == nil {
			k.Close()
			return true
		}
	}
	return false
}
