# Dashboard Capturing Status + Private Server Label

## Goal

Make the dashboard's status badge reflect real traffic instead of process
liveness, and show "Private" as the server name when the user has
configured a custom (non-Albion-Data-Project) ingest destination.

## Background

The dashboard's status badge (`frontend/src/lib/StatusHeader.svelte`)
currently reads `status.CaptureRunning`, which is set `true` the instant
the packet-watcher goroutine starts (`client/albion_watcher.go`'s `run()`)
— not when the client is actually seeing decoded game traffic. This was
already flagged as a known limitation in an earlier review ("the header
reads 'Capturing' whenever the client is up, whether or not the game is
running"). The user now wants the badge to reflect live traffic: start in
a neutral "Ready" state, flip to "Capturing" the moment a real message is
decoded, and fall back to "Ready" after 30 seconds of silence.

Separately, the "Server" field shows West/East/Europe/Unknown based on
`Status.ServerID`, detected from the game server's IP. This has no way to
represent a user who has redirected their public ingest uploads to a
self-hosted server via the `-i` flag (`ConfigGlobal.PublicIngestBaseUrls`)
instead of the default Albion Data Project endpoint — that's an orthogonal
piece of information the frontend currently has no way to see at all.

## Scope

- In scope: activity-based capturing status (backend heartbeat + frontend
  badge), private-ingest detection and display.
- Out of scope: any change to actual packet capture/decoding logic, to the
  upload/ingest pipeline itself, or to any other part of the dashboard.

## Design

### Capturing status: activity heartbeat

"Reliable message" is interpreted broadly: any successfully decoded
Photon `OperationRequest`/`OperationResponse`/`EventData`, regardless of
whether the underlying Photon command was sent reliably or unreliably —
both transport types already flow through the same decode path in
`client/photon/parser.go`, and what the user actually cares about is "are
we seeing real game traffic," not the transport-level distinction.

`internal/dashboard` gets a new activity tracker (alongside the existing
`status.go`/`counters.go`/`loghook.go` files, same package):

- `RecordActivity()` — called from `client/listener.go`'s shared
  `dispatchOperation` method, the single funnel all three parser callbacks
  (`onRequest`, `onResponse`, `onEvent`) already go through, only when
  `err == nil` (i.e. only on a successfully decoded message, not on a
  decode failure). Records the current time and, if `Status.CaptureRunning`
  isn't already `true`, sets it via the existing `setStatus`-style
  change-detected path so an event fires immediately on the transition to
  capturing.
- A background staleness ticker, started via `sync.Once` on the first
  `RecordActivity()` call (mirroring `counters.go`'s coalescing-loop
  pattern), checks periodically whether more than 30 seconds have elapsed
  since the last recorded activity; if so, and `CaptureRunning` is
  currently `true`, it flips back to `false` through the same
  change-detected path.

`client/albion_watcher.go`'s existing `dashboard.SetCaptureRunning(true)` /
`dashboard.SetCaptureRunning(false)` calls (in `run()` and
`closeWatcher()`) are removed — they represented "the watcher goroutine is
alive," a different and less useful signal than "we're seeing real
traffic," and having both would just mean the second write silently wins
over the first depending on timing. `SetCaptureRunning` itself (the
existing setter in `status.go`) stays as the single place that mutates the
field; it's just no longer called from `albion_watcher.go`.

### Private-ingest detection

The frontend never sees `ConfigGlobal.PublicIngestBaseUrls` (the raw `-i`
flag value) — only the backend does, and only the backend can compute
whether it's been overridden. `Status` gets one new field:

```go
PrivateIngest bool
```

Set once during startup (in `startUpdater()` in `albiondata-client.go`,
alongside the existing `dashboard.SetVersionInfo(version, "")` call, since
that's already the "record static startup facts into Status" spot) via a
new setter `dashboard.SetPrivateIngest(private bool)`, computed as:

```go
client.ConfigGlobal.PublicIngestBaseUrls != "https+pow://albion-online-data.com"
```

(the exact string that's `-i`'s documented default in `client/config.go`).
This value never changes after startup — the flag is parsed once at
process start — so a single startup call is sufficient; no need to
recompute it on every status change.

### Frontend (`StatusHeader.svelte`)

- Badge text/color: `CaptureRunning == false` → "Ready", light blue
  background; `CaptureRunning == true` → "Capturing", green background
  (unchanged from today). This replaces the current "Idle"/gray default.
- Server field: `status.PrivateIngest` true → show "Private"; otherwise
  the existing `serverNames[status.ServerID] ?? status.ServerID` lookup,
  unchanged.

## Testing

- `internal/dashboard`: unit tests for the new activity tracker — that
  `RecordActivity()` sets `CaptureRunning` true and emits a status change
  on the false→true transition (and does *not* re-emit on a second call
  while already true, matching the existing change-detection behavior of
  every other setter in this package); that the staleness ticker flips
  `CaptureRunning` back to false after the timeout with no further
  activity (using an injectable/overridable interval the same way
  `counters.go`'s `CoalesceInterval` is overridable in tests, so this
  doesn't require an actual 30-second wait in the test suite).
- `dashboard.SetPrivateIngest` gets the same kind of change-detected-setter
  test as the package's other setters (e.g. `TestSetServer_EmitsOnChange`).
- No test needed for the `client/albion_watcher.go` removal beyond
  confirming the existing `client` package test suite still passes — it's
  a deletion, not new behavior.
- Frontend: no new automated tests (matches this project's existing
  approach — the dashboard UI has none). Manual verification: run the
  client against real or recorded Albion traffic, confirm the badge starts
  "Ready" (light blue), flips to "Capturing" (green) on the first decoded
  message, and reverts to "Ready" after ~30 seconds of no traffic. Confirm
  the Server field shows "Private" when launched with a custom `-i` value,
  and the normal West/East/Europe/Unknown behavior when launched with the
  default.

## Final review fixes

Two corrections landed after the design above was implemented, in a final
whole-branch review pass:

- `Status.PrivateIngest` / `SetPrivateIngest` were renamed to
  `CustomPublicIngest` / `SetCustomPublicIngest`. The old name collided
  confusingly with the codebase's pre-existing, unrelated
  `ConfigGlobal.PrivateIngestBaseUrls` (the `-p` flag, a separate feature
  for uploading a private *copy* of data to a second destination) — the
  new name makes clear this field is about a custom `-i` public-ingest
  URL, not that other feature.
- A new `Status.CaptureError` field and `dashboard.SetCaptureError(bool)`
  setter were added, set to `true` from `runClient()` in
  `albiondata-client.go` when packet capture stops due to a fatal error.
  This gives the dashboard badge a third, distinct "Error" state (red),
  so a broken/dead capture no longer looks identical to a healthy idle
  client — both previously showed "Ready" in light blue.
