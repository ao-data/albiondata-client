# Wails v3 + Svelte Live Status Dashboard

## Goal

Replace the existing `getlantern/systray`-based tray icon with a single-process
Wails v3 (beta) application that provides both the tray icon and a live
status/logs dashboard window, built with a Svelte + Vite frontend.

## Background

The client currently runs packet capture (`runClient()`) on either the main
thread or a goroutine depending on platform, with `systray.Run()` owning the
Cocoa main thread on macOS (required by that platform for any UI). Wails'
webview also requires the main thread on macOS, so running `getlantern/systray`
and Wails in the same process side by side is a known source of conflicts.

Wails v3 solves this by owning the tray icon itself (`systemtray` package,
part of its menu system), cross-platform on Windows, macOS, and Linux — with
the caveat that on Linux, without a notification-area extension present, the
tray icon simply won't appear (a pre-existing limitation of Linux desktop
environments, not something Wails or this project can work around).

## Scope

- In scope: tray icon + dashboard window via Wails v3, live status/counters/log
  display, removal of `getlantern/systray`.
- Out of scope: editing `config.yaml` from the UI, browsing captured market
  data itself, any change to packet capture logic.

## Architecture

Single Go binary. Wails v3 becomes the process owner: `app.Run()` runs on the
main thread and owns both the tray icon and the dashboard window. Packet
capture (`runClient()`) keeps running in its own goroutine exactly as today —
capture logic is untouched and does not depend on whether the dashboard
window is open.

The dashboard window is created hidden at startup. The tray's "Open
Dashboard" menu item shows it; closing the window hides it (the app keeps
running), matching the current tray app's behavior. The tray also keeps
"Open Log File" and "Quit" items from the existing implementation.

The existing `systray/` package (all three platform files) is removed and
replaced by tray setup living alongside the Wails app bootstrap in
`albiondata-client.go`.

## Data flow (backend → frontend)

Everything is push-based via Wails events — no frontend polling.

1. **Status** (`status:changed` event) — capture running state, detected
   server (`albionState.AODataServerID` / `AODataIngestBaseURL`), client
   version, and update-available flag (from the existing updater). Emitted
   whenever any of these values change.

2. **Upload counters** (`counters:snapshot` event) — new thread-safe counters
   (atomic ints), one per upload type (market orders, gold prices, market
   histories, festivities, map data), incremented at each existing
   `sendMsgToPublicUploaders` / `sendToIngest` call site. A ticker checks for
   changes and emits a single coalesced snapshot at most every ~500ms, to
   avoid flooding the frontend during high-volume capture bursts. No event
   fires if nothing changed since the last tick.

3. **Live log tail** (`log:line` event) — a `logrus.Hook` registered via the
   existing `log.AddHook()` keeps a capped ring buffer (last ~500 lines) and
   emits one event per new line as it's logged. A bound `GetRecentLogs()`
   method backfills the buffer when the window is shown; after that the
   frontend only listens for `log:line` events. The ring buffer keeps
   running regardless of window visibility, so no lines are lost between
   opens.

New backend code lives in small, focused packages:
- `internal/dashboard/status.go` — status snapshot + change detection
- `internal/dashboard/counters.go` — atomic counters + coalescing ticker
- `internal/dashboard/loghook.go` — logrus hook + ring buffer

Each is testable in isolation without a running Wails app or capture session.

## Frontend

Svelte + Vite, scaffolded from Wails' official `svelte` template (not
SvelteKit — its SSR/routing model doesn't fit a desktop webview, and Wails
does not ship a SvelteKit template). Single page, no client-side routing.

Layout:
- **Header** — connection status badge (capturing/idle), detected server,
  version + update-available badge
- **Counters panel** — live tallies per upload type
- **Log tail** — scrolling, capped list of recent log lines, color-coded by
  level, auto-scrolls unless the user has scrolled up

All three sections subscribe to the Wails events described above.

## Error handling

- Tray creation follows Wails v3 guidance: check the platform, create the
  tray, and degrade gracefully if it doesn't appear (Linux without a
  notification-area extension). This is documented as a known platform
  limitation — no fallback CLI flag or alternate entry point is added.
- No new error handling is needed around capture logic; it is unchanged.

## Testing

- Go unit tests (table-driven, matching the existing style in
  `client/albion_state_test.go`) for:
  - counter increment and coalescing/change-detection logic
  - log ring buffer capping and line formatting
- No frontend test framework is added — this is a status display, not
  business logic (YAGNI).
- Manual verification: run `wails3 dev` and exercise capture status,
  counter increments, and log tail streaming against a live or simulated
  capture session before considering the work complete.

## Dependencies

- Add: Wails v3 (beta) CLI/runtime
- Remove: `github.com/getlantern/systray`, `systray/` package
