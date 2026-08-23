# Dashboard Auto-Open + Windows Console Hide/Show

## Goal

Two small, related enhancements to the Wails v3 dashboard app added in
[2026-08-08-wails-svelte-dashboard-design.md](2026-08-08-wails-svelte-dashboard-design.md):

1. The dashboard window opens automatically at launch, instead of starting
   hidden and requiring a tray click.
2. On Windows, the console window is hidden by default at launch (when
   launched with no arguments), with a tray menu item to reveal it again.

## Background

The dashboard currently starts with `Hidden: true` and is only shown via
the tray's "Open Dashboard" item or a fatal capture error. Separately, the
pre-Wails-v3 tray implementation had a Windows-only "Show/Hide Console"
toggle (via `gonutz/w32`, since removed) that was dropped when the tray was
rewritten on Wails v3, on the reasoning that the new dashboard's log tail
was a better replacement. That reasoning still holds for *viewing logs*,
but the user now wants the console hidden by default (rather than shown by
default, as it currently is on Windows) with an escape hatch to reveal it
when needed — e.g. for debugging output the dashboard doesn't surface.

The user's longer-term direction (not in scope for this work, noted for
context): eventually move toward a "double-click the desktop icon to open
the dashboard" launch model on all platforms, while still supporting
command-line flags for advanced use; a proper installer/updater on macOS;
and a Snap package on Linux. This spec's console-hiding behavior is
designed to not conflict with that direction — flag usage remains visible.

## Scope

- In scope: auto-opening the dashboard window at launch; hiding/showing
  the Windows console via a new `internal/console` package and a tray
  toggle.
- Out of scope: any macOS/Linux equivalent of console hiding (console
  windows aren't a comparable concept there — see the design discussion
  in the parent spec's tray section); installer/updater/Snap work; any
  new CLI flags.

## Design

### Auto-open at launch

In `runDashboardApp()` (`albiondata-client.go`), the `WebviewWindowOptions`
passed to `app.Window.NewWithOptions` still set `Hidden: true`. Visibility
and focus instead come entirely from an
`app.Event.OnApplicationEvent(events.Common.ApplicationStarted, ...)` hook
registered later in the same function, which calls the existing
`showDashboardWindow()` helper (`Show()` then `Focus()`, with an
`IsVisible()`-based retry to cover a narrow goroutine-scheduling race in
window realization). This was chosen over a bare `Hidden: false` plus an
immediate `Focus()` call after `NewWithOptions()`: window realization
during Wails' startup happens asynchronously in its own goroutine, so
calling `Focus()` right after creation would silently no-op (the window
isn't realized yet). `ApplicationStarted` is guaranteed to fire only after
Wails' internal startup, including window realization, has completed.
Keeping `Hidden: true` also avoids an initial-paint/flash on Windows, since
the native window is never marked visible until content is ready, rather
than being shown and then immediately repainted. `Focus()` matters because
the app runs under `ActivationPolicyAccessory` on macOS, where window
visibility alone doesn't activate the app — without it the window could
open behind whatever app currently has focus (the same class of bug fixed
for the tray's "Open Dashboard" click). All existing behavior is
unchanged: closing the window still hides it rather than quitting, and the
tray's "Open Dashboard" item still works the same way for reopening it
later.

### Windows console hide/show

A new package, `internal/console`, split by build tag the same way
`icon/` already is for platform-specific assets:

- `internal/console/console_windows.go` (`//go:build windows`) — the real
  implementation, using `syscall.NewLazyDLL("kernel32.dll")` to resolve
  `GetConsoleWindow` and `syscall.NewLazyDLL("user32.dll")` to resolve
  `ShowWindow`. This is standard-library only (`syscall`), replacing what
  `gonutz/w32` used to provide without reintroducing that dependency.
- `internal/console/console_other.go` (`//go:build !windows`) — no-op
  stubs, so callers never need `runtime.GOOS` checks around calling into
  this package.

Public API:

```go
func Supported() bool // true only on windows
func Owned() bool
func Hide()
func Show()
func Hidden() bool
```

`Owned()` reports whether this process is the only one attached to its
console, via `kernel32.GetConsoleProcessList`. This matters because
`GetConsoleWindow()` returns the *inherited* console when launched from an
already-open shell, not a private one — without this check, a no-argument
launch from an open PowerShell/cmd window would hide that shell itself,
with no way to restore it. `Owned()` is true for a fresh Explorer/shortcut
launch (Windows allocates a private console) and false when sharing a
parent shell's console.

In `main()` (`albiondata-client.go`), before any other startup work: if
`runtime.GOOS == "windows" && len(os.Args) == 1 && console.Owned()`, call
`console.Hide()` immediately, so nothing flashes on screen before it
disappears. Launching with any argument (`-version`, `-h`, `-debug`, etc.)
leaves the console visible, so flag output is never silently swallowed —
this is the reason for gating on argument count rather than hiding
unconditionally. The `Owned()` check is what prevents hiding a console
this process doesn't own.

In `setupTray()`, only when `console.Supported()` is true (i.e. only on
Windows), add one toggle menu item to the tray menu. It starts labeled
"Show Console" (since the console starts hidden on a no-argument launch;
if launched with arguments the console was never hidden, so the toggle's
initial label reflects `console.Hidden()`'s actual state at tray setup
time, not a hardcoded assumption). Its `OnClick` handler calls
`console.Show()`/`console.Hide()` based on current state and relabels
itself via `ctx.ClickedMenuItem().SetLabel(...)`, mirroring the toggle
pattern the pre-Wails-v3 Windows tray used. The item is not added at all
on macOS/Linux.

The existing `client.ConfigGlobal.Minimize` flag remains an unused no-op,
as already documented in the parent spec — this work does not revive or
repurpose it.

## Testing

- `internal/console` gets no unit tests: `console_windows.go` wraps opaque
  WinAPI calls with no meaningful behavior to assert from a non-Windows
  CI/dev machine, and `console_other.go`'s stubs are trivial enough that a
  test would only restate their bodies. This matches the project's
  existing approach to platform-specific, unverifiable-in-CI code (e.g.
  `icon/iconwin.go` has no test either).
- Manual verification: build and run on Windows, confirm (a) no-argument
  launch hides the console immediately with no visible flash, (b) the tray
  toggle shows/hides it and relabels correctly both directions, (c)
  launching with `-version` or `-h` leaves the console visible and its
  output is readable. Build and run on macOS/Linux, confirm the dashboard
  auto-opens and focuses at launch, and that no console-related tray item
  appears.
