# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A third-party client that passively captures local network traffic,
decodes Albion Online's Photon protocol packets, and ships relevant
data (market orders, gold prices, etc.) to a central NATS ingest server
that anyone can subscribe to. It monitors only - it never modifies or
injects traffic. Ships as a single executable per OS (Windows, macOS,
Linux), with a Wails v3 desktop dashboard plus a system tray icon.

## Commands

```
make run              # go run the client directly
make frontend         # build the Svelte dashboard (npm ci && npm run build)
make fmt               # goimports -w across the repo
make validate-fmt      # goimports -l check, non-zero exit on violations
make build-windows     # scripts/build-windows.sh
make build-linux       # scripts/build-linux.sh
make build-darwin      # scripts/build-darwin.sh

go build ./...          # needs `make frontend` first - see build-and-release.md
go vet ./...
go test ./...
go test ./client/... -run TestName -v   # single test
```

**Before running `go build`/`go vet`/`go test` on the root package**,
run `make frontend` at least once - `albiondata-client.go` has
`//go:embed all:frontend/dist`, and that fails at compile time if the
directory doesn't exist yet, even for commands that don't touch the
embedded FS.

**Never add a new `.go` file directly in the repo root** expecting
`main` to pick it up - every build script invokes `go build
albiondata-client.go` (the single file, not `.`/the package), so
anything new that `main` needs must live in an `internal/` package and
be imported normally. Full detail in `docs/build-and-release.md`.

## Architecture

**Capture pipeline** (`client/`): `gopacket/pcap` captures raw UDP
packets -> `client/photon` parses the Photon protocol framing -> a
`Router` (`router.go`) dispatches decoded requests/responses/events to
per-type handlers. Each Photon operation/event has its own file
(`operation_*.go`, `event_*.go`) implementing `Process(state
*albionState)` against shared client state (`albion_state.go`) -
that's the pattern to follow for handling a new operation/event type,
not a big switch statement. Uploaders (`uploader_nats.go`,
`uploader_http.go`, `uploader_http_pow.go`) push processed data to the
ingest servers; which server is chosen is resolved from the observed
game server IP in `albionState.GetServer()`.

**Dashboard** (`internal/dashboard/`): a plain state-holder package,
independent of Wails, that the capture pipeline and the GUI both talk
to - setters like `SetCaptureRunning`/`SetVersionInfo` push changes
out via a single registered callback, the GUI layer wires that callback
to the Wails frontend binding. Full detail in `docs/gui-dashboard.md`.

**GUI** (`frontend/`, Svelte 5 + Vite, Wails v3): native webview window
plus system tray, built as a plain Go binary (no separate Wails CLI
build step - just `npm run build` then `go build`). Full detail in
`docs/gui-dashboard.md`.

**Windows packet-capture driver** (`internal/pcapdriver/`): detects
whether Npcap or the old, abandoned WinPcap is installed and warns the
user on startup if capture won't work. Full history and the VPN
adapter bug this all stems from: `docs/windows-capture-pcap.md`.

**Self-update**: hourly poll against a GitHub repo's releases,
replacing the running binary in place. Exact-match filename logic and
a real arm64/amd64 gotcha it produces: `docs/self-update.md`.

**Build, CI, and release**: reusable GitHub Actions workflow, the
single-file build convention, and Windows/macOS build specifics:
`docs/build-and-release.md`.

## Further reading

- [docs/gui-dashboard.md](docs/gui-dashboard.md) - Wails v3 + Svelte
  dashboard architecture, design decisions, why there's no macOS `.app`
  bundle.
- [docs/windows-capture-pcap.md](docs/windows-capture-pcap.md) -
  WinPcap -> Npcap migration, the driver-detection state machine, and
  the Wintun/PIA VPN adapter capture bug.
- [docs/build-and-release.md](docs/build-and-release.md) - build
  scripts, the reusable CI workflow, Windows/macOS build specifics,
  what got deleted and why.
- [docs/self-update.md](docs/self-update.md) - updater filename
  matching, hardcoded target repo, the arm64 local-build gotcha.
