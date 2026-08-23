# Dashboard Auto-Open + Windows Console Hide/Show Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Open the dashboard window automatically (and focused) at launch, and hide the Windows console by default on a no-argument launch with a tray toggle to reveal it.

**Architecture:** A new leaf package, `internal/console`, wraps the two WinAPI calls needed to hide/show the process's console window (build-tag-split, matching the existing `icon/` package's pattern), with no-op stubs on non-Windows. `albiondata-client.go` calls into it conditionally and adds one tray menu item on Windows only. The dashboard's auto-open reuses the existing `showDashboardWindow()` helper (already used for the fatal-capture-error path) rather than relying on the window's `Hidden` option, because window realization during Wails' startup sequence happens asynchronously in its own goroutine — `showDashboardWindow()`'s `Show()` call self-heals that race (it forces synchronous realization if needed via `InvokeSync(w.Run)`), while a bare `Hidden: false` option would not guarantee the window is both visible and focused by the time the app appears to the user.

**Tech Stack:** Go 1.24 (existing), `syscall` (stdlib, for the Windows console APIs — no new dependency), Wails v3 (existing).

## Global Constraints

- `package main` remains exactly one file, `albiondata-client.go`, at the repo root (existing build-script constraint from the parent dashboard plan — unchanged, not touched by this plan since `internal/console` is a separate importable package).
- Console hiding is Windows-only. `internal/console`'s non-Windows build provides no-op stubs so callers never need `runtime.GOOS` checks of their own.
- Console hiding only triggers when launched with no arguments (`len(os.Args) == 1`), so flag output (`-version`, `-h`, etc.) is never silently hidden.
- No new tray items appear on macOS/Linux from this plan — the console toggle is gated on `console.Supported()`, which is `true` only on Windows.
- `client.ConfigGlobal.Minimize` remains an unused no-op, as already established — this plan does not revive it.

---

## File Structure

New:
- `internal/console/console_windows.go` — real console hide/show via `syscall.NewLazyDLL`
- `internal/console/console_other.go` — no-op stubs for all other platforms

Modified:
- `albiondata-client.go` — auto-open the dashboard at launch; hide the console conditionally at startup; add the Windows-only tray toggle

---

### Task 1: `internal/console` package

**Files:**
- Create: `internal/console/console_windows.go`
- Create: `internal/console/console_other.go`

**Interfaces:**
- Produces: `func Supported() bool`, `func Hide()`, `func Show()`, `func Hidden() bool` — all four used by Task 3's changes to `albiondata-client.go`.

No unit tests for this task — see the design spec's Testing section (`docs/superpowers/specs/2026-08-09-dashboard-autoopen-console-hide-design.md`): `console_windows.go` wraps opaque WinAPI calls with no meaningful behavior to assert from a non-Windows dev/CI machine, and `console_other.go`'s stubs are trivial enough that a test would only restate their bodies. This matches the existing project pattern for platform-specific, unverifiable-in-CI code (e.g. `icon/iconwin.go` has no test either). Verification here is cross-compilation, not `go test`.

- [ ] **Step 1: Create the non-Windows stub**

```go
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
```

- [ ] **Step 2: Create the Windows implementation**

```go
// internal/console/console_windows.go
//go:build windows

package console

import "syscall"

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	user32                = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
)

// SW_HIDE and SW_SHOW are the ShowWindow nCmdShow values used here. See
// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-showwindow
const (
	swHide = 0
	swShow = 5
)

var hidden bool

// Supported reports whether console hiding is implemented on this
// platform. Always true on Windows.
func Supported() bool {
	return true
}

// Hide hides the process's console window, if one exists (e.g. none
// exists if the process wasn't allocated a console at all).
func Hide() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swHide)
	hidden = true
}

// Show reveals the process's console window, if one exists.
func Show() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swShow)
	hidden = false
}

// Hidden reports whether Hide has been called more recently than Show.
func Hidden() bool {
	return hidden
}
```

- [ ] **Step 3: Verify both platform variants compile**

Run (from the repo root, on macOS or Linux):
```bash
go build ./internal/console/...
```
Expected: success — this compiles `console_other.go` for the host platform.

Run:
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./internal/console/...
```
Expected: success — this compiles `console_windows.go`. `syscall.NewLazyDLL`/`NewProc`/`.Call` are all pure-Go (no cgo) on the `windows` GOOS, so this cross-compiles cleanly from macOS/Linux with no Windows toolchain needed.

- [ ] **Step 4: Commit**

```bash
git add internal/console/console_windows.go internal/console/console_other.go
git commit -m "feat: add internal/console package for Windows console hide/show"
```

---

### Task 2: Auto-open the dashboard at launch

**Files:**
- Modify: `albiondata-client.go`

**Interfaces:**
- Consumes: `showDashboardWindow()` (already defined at the top of `albiondata-client.go`, calls `w.Show(); w.Focus()` with a nil-safety check).

- [ ] **Step 1: Add the auto-open call in `runDashboardApp()`**

In `albiondata-client.go`, find this block near the end of `runDashboardApp()`:

```go
	setupTray(app, dashboardWindow)

	if err := app.Run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
```

Replace it with:

```go
	setupTray(app, dashboardWindow)

	// Open and focus the dashboard automatically at launch. This runs in
	// its own goroutine because app.Run() below blocks until the app
	// exits — Show()/Focus() are safe to call from any goroutine (they
	// dispatch onto the main thread internally), and Show() forces the
	// window to finish realizing even if Wails' own async startup
	// sequence hasn't gotten to it yet.
	go showDashboardWindow()

	if err := app.Run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
```

This is the only change for this task. The window's `WebviewWindowOptions` keep `Hidden: true` unchanged — visibility comes entirely from this explicit call, the same mechanism already used for the tray's "Open Dashboard" click and the fatal-capture-error path, so there's exactly one code path responsible for "make the dashboard visible and focused" across the whole file.

- [ ] **Step 2: Verify it builds**

```bash
go build ./client/... ./internal/... ./icon/...
cd frontend && npm run build && cd ..
go build albiondata-client.go
```
Expected: all succeed (only pre-existing benign macOS linker SDK-version warnings, if any, on the last command).

- [ ] **Step 3: Manual verification (macOS or Linux — whichever you're on)**

Run the built binary, confirm the dashboard window appears and is focused (in front of other windows) immediately at launch, with no click needed. Confirm the tray's "Open Dashboard" still works afterward if you close the window. Report what you observed — this is a visual/interactive check the automated build steps can't confirm on their own.

- [ ] **Step 4: Commit**

```bash
git add albiondata-client.go
git commit -m "feat: open and focus the dashboard automatically at launch"
```

---

### Task 3: Windows console hide at launch + tray toggle

**Files:**
- Modify: `albiondata-client.go`

**Interfaces:**
- Consumes: `console.Supported() bool`, `console.Hide()`, `console.Show()`, `console.Hidden() bool` (Task 1).

- [ ] **Step 1: Add the import**

In `albiondata-client.go`'s import block, add `"github.com/ao-data/albiondata-client/internal/console"` alongside the other `github.com/ao-data/albiondata-client/...` imports, keeping the block ordered alphabetically by import path (matching the existing style):

```go
	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/icon"
	"github.com/ao-data/albiondata-client/internal/console"
	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/log"
```

- [ ] **Step 2: Hide the console at the very start of `main()`**

Find:

```go
func main() {
	if client.ConfigGlobal.PrintVersion {
```

Replace with:

```go
func main() {
	// Hide the console immediately, before any other startup work, so
	// nothing flashes on screen before it disappears. Only when launched
	// with no arguments — any flag (e.g. -version, -h) means the user is
	// running this from a terminal and expects to see output, so leave
	// the console visible in that case.
	if runtime.GOOS == "windows" && len(os.Args) == 1 {
		console.Hide()
	}

	if client.ConfigGlobal.PrintVersion {
```

- [ ] **Step 3: Add the tray toggle item**

In `setupTray()`, find:

```go
	menu.Add("Open Log File").OnClick(func(ctx *application.Context) {
		openLogFile()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
```

Replace with:

```go
	menu.Add("Open Log File").OnClick(func(ctx *application.Context) {
		openLogFile()
	})

	if console.Supported() {
		consoleLabel := "Show Console"
		if !console.Hidden() {
			consoleLabel = "Hide Console"
		}
		menu.Add(consoleLabel).OnClick(func(ctx *application.Context) {
			item := ctx.ClickedMenuItem()
			if console.Hidden() {
				console.Show()
				item.SetLabel("Hide Console")
			} else {
				console.Hide()
				item.SetLabel("Show Console")
			}
		})
	}

	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
```

The item's starting label reflects `console.Hidden()`'s actual state at tray setup time (which runs after `main()`'s conditional `console.Hide()` call above), not a hardcoded assumption — so a launch with arguments (console never hidden) correctly starts the item at "Hide Console", not "Show Console".

- [ ] **Step 4: Verify it builds on every platform**

```bash
go build ./client/... ./internal/... ./icon/...
cd frontend && npm run build && cd ..
go build albiondata-client.go
```
Expected: success (this exercises the `!windows` build of `internal/console` via whatever platform you're building on).

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/albiondata-client-windows-check.exe albiondata-client.go
```
Expected: success — this exercises the `windows` build of `internal/console` together with the rest of `albiondata-client.go`, cross-compiled. Delete the throwaway binary afterward: `rm -f /tmp/albiondata-client-windows-check.exe`.

- [ ] **Step 5: Manual verification**

On whatever platform you're on (macOS/Linux): confirm no console-related tray item appears (since `console.Supported()` is false there), and everything else about the tray/dashboard still works as before.

On Windows, if available: confirm (a) launching with no arguments hides the console immediately with no visible flash, (b) the tray item starts labeled "Show Console" and toggles correctly in both directions, relabeling itself each click, (c) launching with `-version` or another flag from a terminal leaves the console visible with its output readable, and the tray item in that case starts labeled "Hide Console". If you don't have a Windows machine to test on, say so explicitly in your report rather than claiming this was verified — the cross-compile check in Step 4 only proves it builds, not that it behaves correctly at runtime.

- [ ] **Step 6: Commit**

```bash
git add albiondata-client.go
git commit -m "feat: hide Windows console on launch with a tray toggle to reveal it"
```
