# Windows capture: WinPcap, Npcap, and the VPN adapter bug

Packet capture on Windows goes through `gopacket/pcap`, which links
against `wpcap.dll` via cgo (`thirdparty/WpdPack/` holds the headers +
import libraries that link is built against). What provides that DLL
at runtime is the part that changed.

## WinPcap -> Npcap

The installer used to bundle `thirdparty/WinPcap_4_1_3.exe` (from
2013) and silently install it for every Windows user. This turned out
to be broken for a large chunk of users: WinPcap can't open
Wintun/WireGuard-style virtual adapters at all (pure Layer-3, no
Ethernet framing - WinPcap predates their existence by a decade), and
is also unreliable on modern Windows 10/11 NDIS stacks even for
physical adapters when a VPN's own filter driver sits in the chain -
which is exactly the setup a lot of users have (PIA, WireGuard, etc.).
Symptom: capture opens with no error, sees zero packets.

Fix was to stop bundling a driver at all and point users at
[Npcap](https://npcap.com) instead - **not** a drop-in bundle
replacement, because Npcap redistribution requires a paid OEM license
(the Nmap Project's own guidance for free/open-source software is:
don't bundle, detect its absence and link out). Two places now do
that detection, because a bundled installer's check only ever catches
a *fresh install*, not an existing user who already has old WinPcap
and never re-runs the installer:

1. **Install time**: `pkg/nsis/albiondata-client.nsi`'s `.onInit`
   checks `HKLM\SOFTWARE\Npcap` (+ the `WOW6432Node` fallback for
   32-bit-on-64-bit) via `ReadRegDWORD`/`IfErrors`, and offers to open
   `https://npcap.com/#download` if it's missing. Doesn't block the
   install either way, just warns.
2. **Every startup**: `internal/pcapdriver` (Windows-only,
   `checkCaptureDriver()` in `albiondata-client.go`, called from
   `main()` right after the log hook is set up). This is the one that
   actually matters for existing installs.

Npcap installs in "WinPcap API-compatible mode" by default, so no
change was needed to `thirdparty/WpdPack/` or anything `gopacket/pcap`
links against - same `wpcap.dll` export surface.

## internal/pcapdriver: the four states

`Check()` (real implementation in `pcapdriver_windows.go`, no-op stub
in `pcapdriver_other.go` for every other OS) reads three things -
Npcap's registry key and its `AdminOnly` DWORD, legacy WinPcap's
registry key, and whether this process is running elevated - and
`decide()` (a separate pure function, specifically so it's unit
testable without touching real registry state - see
`pcapdriver_windows_test.go`) turns that into one of four outcomes:

1. Npcap present, all-users mode -> no warning.
2. Npcap present, admin-only mode, process elevated -> no warning
   (capture works as long as the client stays elevated).
3. Npcap present, admin-only mode, process **not** elevated -> warns,
   tells the user to restart as Administrator. No `HelpURL` - a
   download link isn't the fix here, restarting elevated is.
4. Npcap absent (whether or not old WinPcap is present) -> warns with
   an npcap.com `HelpURL`. Message text distinguishes "only outdated
   WinPcap was detected" from "nothing at all is installed" - they're
   different problems for the user to reason about.

The all-users-vs-admin-only tradeoff is stated to the user explicitly
rather than picked silently: all-users lets this client capture without
elevation, but means any other app on the system could also read
traffic without admin rights either; admin-only avoids that but means
this client must be launched elevated every time. See the
`npcapGuidance` string in `pcapdriver_windows.go` for the exact wording
if this ever needs to change.

Surfacing is both a log line (via the existing log hook, so it lands in
the dashboard's LogTail automatically) and a dedicated banner -
`Status.DriverWarning`/`DriverHelpURL` in `internal/dashboard`, rendered
by `Sidebar.svelte` above the counters, with a "Get Npcap" link when
`DriverHelpURL` is set (opened via `Browser.OpenURL`, matching
`Footer.svelte`'s external links). Log-only was considered and
rejected - a log-only line is exactly what made the original bug take
multiple sessions to diagnose the first time around.

## The Wintun/PIA adapter bug (root cause, now fixed)

Even after moving to Npcap, one specific case stayed broken: capture
with a VPN like PIA active never saw any Albion traffic, even though
Npcap could open adapters without error. Two separate bugs stacked:

1. **`client/net_interface_filter_win.go`**: `physicalAddrToString()`
   formatted the full fixed `[8]byte` `PhysicalAddress` array
   unconditionally, ignoring the separate `PhysicalAddressLength`
   field. Adapters with no real MAC - Wintun, PIA's default Windows VPN
   driver, is pure Layer-3 with no Ethernet framing - report
   `PhysicalAddressLength=0` with a zeroed array, which formatted to
   `"00:00:00:00:00:00:00:00"`. That string happened to match the
   `"00:00:00:00:00"` Teredo-pseudo-interface entry in
   `macAddrPartsToFilter` (`net_interface_filter.go`), so the PIA
   tunnel adapter was silently excluded from the capture candidate
   list entirely. Fix: respect `PhysicalAddressLength` (zero-length ->
   empty string, which correctly passes the filter instead of
   colliding with it).
2. **`client/listener.go`**: once fix #1 made the Wintun adapter a
   capture candidate, *opening* it still failed - Npcap can't bind to a
   pure Layer-3 Wintun adapter at all, which is an OS/driver limitation,
   not fixable here. That's expected and fine for that one adapter, but
   the failure path called `log.Panic(err)`, and `startOnline` runs on
   its own goroutine (`albion_watcher.go`'s `createListeners`) with no
   `recover()` - so the panic crashed the *entire process*, killing
   capture on every other (working) interface too, not just the one
   that couldn't open. Fixed: failures now log via `log.Errorf` and
   abandon just that one listener; `stop()` was made nil-safe for
   listeners that never got a handle.

Both were verified by cross-compiling for `windows/amd64` and running
the actual Windows test suite, since a lot of this debugging happened
without a Windows machine on hand (`GOOS=windows GOARCH=amd64
CGO_ENABLED=0 go build/test ./client/...` - no cgo needed for these
particular files).

Traffic itself is Photon protocol over UDP port 5056 - if a report of
"no traffic" ever comes back after all of the above, check whether
Albion is actually logged in with the market screen open (just running
the game isn't enough to generate traffic to see).
