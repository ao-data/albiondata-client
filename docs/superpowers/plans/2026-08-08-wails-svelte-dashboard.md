# Wails v3 + Svelte Live Status Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `getlantern/systray` with a single-process Wails v3 (beta) application that owns the tray icon and provides a live status/counters/log dashboard window, built with a Svelte + Vite frontend.

**Architecture:** `internal/dashboard` is a new leaf package holding all dashboard state (status, upload counters, log ring buffer) with change-detection so it only pushes updates when something actually changed. Existing `client` package call sites call into it (a handful of one-line additions). `albiondata-client.go` (the sole file in `package main`) is rewritten to bootstrap a Wails v3 `application.App`, which owns the tray icon and a hidden-by-default dashboard window on the single main thread, and forwards `internal/dashboard` state changes to the frontend as Wails events. `runClient()` (packet capture) keeps running in its own goroutine, unaffected by whether the dashboard window is open.

**Tech Stack:** Go 1.24 (existing), Wails v3 beta (`github.com/wailsapp/wails/v3`), Svelte + Vite (`@wailsio/runtime` for the JS/Svelte side), logrus (existing, via a new hook).

## Global Constraints

- Single Go binary; no separate GUI process. (Design: single-process architecture.)
- Wails v3's `systemtray` package fully replaces `getlantern/systray` — the `systray/` package is deleted, `github.com/getlantern/systray` and `github.com/gonutz/w32` are removed from `go.mod`.
- No polling from the frontend — all of status, counters, and logs are pushed via Wails events. (Design: data flow section.)
- Upload counter events are coalesced to at most one snapshot per ~500ms, only when a counter actually changed. (Design: counter events.)
- `package main` remains exactly one file, `albiondata-client.go`, at the repo root — this is required by the existing build scripts (`scripts/build-*.sh`), which invoke `go build ... albiondata-client.go` naming that file explicitly. All new logic goes in importable packages; only entrypoint wiring goes in that file.
- The project vendors its dependencies (`vendor/` is checked in); every dependency change must be followed by `go mod tidy && go mod vendor`.
- On Linux, if no tray/notification-area extension is present, the tray icon simply won't appear — this is a known, undocumented-workaround platform limitation (Design: error handling). No fallback entry point is added for it.
- The existing Windows-only "Show/Hide Console" tray item is dropped. It's superseded by the new dashboard's live log tail, which is strictly more useful than toggling a console window, and isn't part of the approved scope. `client.ConfigGlobal.Minimize` (the flag that drove it) is left defined but becomes a no-op — Wails windows already start hidden by default, so the flag's original intent ("don't pop up a window") is preserved by default behavior.
- No config-editing UI, no market-data browsing UI — out of scope (Design: scope).

---

## File Structure

New:
- `internal/dashboard/status.go` — capture/server/version status snapshot, change-detected setters
- `internal/dashboard/status_test.go`
- `internal/dashboard/counters.go` — atomic per-topic upload counters, coalesced change notification
- `internal/dashboard/counters_test.go`
- `internal/dashboard/loghook.go` — logrus hook + capped ring buffer of recent log lines
- `internal/dashboard/loghook_test.go`
- `internal/dashboard/service.go` — `DashboardService`, the Wails-bound struct the frontend calls for backfill
- `internal/dashboard/service_test.go`
- `icon/tray.go` — cross-platform (no build tag) PNG icon bytes for the Wails tray, reusing the existing artwork
- `frontend/` — Vite + Svelte app (scaffolded fresh; the existing untracked `frontend/dist` and `frontend/node_modules` are stale leftovers from an earlier abandoned attempt and are deleted first)

Modified:
- `client/listener.go` — report detected server to `internal/dashboard` after `GetServer()`
- `client/albion_watcher.go` — report capture running/stopped to `internal/dashboard`
- `client/dispatcher.go` — increment the upload counter in `sendMsgToPublicUploaders`
- `albiondata-client.go` — rewritten entrypoint: Wails v3 bootstrap, tray menu, hidden dashboard window, event wiring
- `go.mod`, `go.sum`, `vendor/` — remove `getlantern/systray`, `gonutz/w32`; add `github.com/wailsapp/wails/v3`
- `scripts/build-darwin.sh`, `scripts/build-linux.sh`, `scripts/build-windows.sh`, `scripts/run.sh` — build the frontend before building Go
- `Makefile` — add a `frontend` target

Deleted:
- `systray/systray_darwin.go`, `systray/systray_win.go`, `systray/systray_others.go`

---

### Task 1: Dashboard status tracking

**Files:**
- Create: `internal/dashboard/status.go`
- Test: `internal/dashboard/status_test.go`

**Interfaces:**
- Produces: `type Status struct { Version, UpdateAvailable string; CaptureRunning bool; ServerID int; IngestBaseURL string }`, `func GetStatus() Status`, `func SetVersionInfo(version, updateAvailable string)`, `func SetCaptureRunning(running bool)`, `func SetServer(serverID int, ingestBaseURL string)`, `func OnStatusChange(fn func(Status))` — all used by later tasks.

- [ ] **Step 1: Write the failing tests**

```go
// internal/dashboard/status_test.go
package dashboard

import "testing"

func TestSetServer_EmitsOnChange(t *testing.T) {
	resetStatusForTest()

	var got []Status
	OnStatusChange(func(s Status) { got = append(got, s) })

	SetServer(1, "https+pow://pow.west.albion-online-data.com")
	SetServer(1, "https+pow://pow.west.albion-online-data.com") // no change, no emit
	SetServer(2, "https+pow://pow.east.albion-online-data.com")

	if len(got) != 2 {
		t.Fatalf("expected 2 emits, got %d: %+v", len(got), got)
	}
	if got[0].ServerID != 1 || got[1].ServerID != 2 {
		t.Fatalf("unexpected emitted statuses: %+v", got)
	}
}

func TestGetStatus_ReflectsAllSetters(t *testing.T) {
	resetStatusForTest()

	SetVersionInfo("1.2.3", "1.3.0")
	SetCaptureRunning(true)
	SetServer(3, "https+pow://pow.europe.albion-online-data.com")

	got := GetStatus()
	want := Status{
		Version:         "1.2.3",
		UpdateAvailable: "1.3.0",
		CaptureRunning:  true,
		ServerID:        3,
		IngestBaseURL:   "https+pow://pow.europe.albion-online-data.com",
	}
	if got != want {
		t.Fatalf("GetStatus() = %+v, want %+v", got, want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/dashboard/... -run TestSetServer_EmitsOnChange -v`
Expected: FAIL — `resetStatusForTest`, `OnStatusChange`, `SetServer`, `Status` undefined (package doesn't exist yet).

- [ ] **Step 3: Implement**

```go
// internal/dashboard/status.go
package dashboard

import "sync"

// Status is a point-in-time snapshot of the client's observable state,
// pushed to the dashboard frontend whenever it changes.
type Status struct {
	Version         string
	UpdateAvailable string
	CaptureRunning  bool
	ServerID        int
	IngestBaseURL   string
}

var (
	statusMu   sync.Mutex
	status     Status
	statusEmit func(Status)
)

// OnStatusChange registers the callback invoked whenever the status
// changes. Only one callback is supported; a later call replaces the
// earlier one.
func OnStatusChange(fn func(Status)) {
	statusMu.Lock()
	statusEmit = fn
	statusMu.Unlock()
}

// GetStatus returns the current status snapshot.
func GetStatus() Status {
	statusMu.Lock()
	defer statusMu.Unlock()
	return status
}

func setStatus(next Status) {
	statusMu.Lock()
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// SetVersionInfo records the running client version and, if one is known,
// the version available to update to (empty string if none).
func SetVersionInfo(version, updateAvailable string) {
	next := GetStatus()
	next.Version = version
	next.UpdateAvailable = updateAvailable
	setStatus(next)
}

// SetCaptureRunning records whether packet capture is currently active.
func SetCaptureRunning(running bool) {
	next := GetStatus()
	next.CaptureRunning = running
	setStatus(next)
}

// SetServer records the Albion Online server detected from captured
// traffic and the ingest URL that maps to it.
func SetServer(serverID int, ingestBaseURL string) {
	next := GetStatus()
	next.ServerID = serverID
	next.IngestBaseURL = ingestBaseURL
	setStatus(next)
}

// resetStatusForTest clears all package state. Test-only.
func resetStatusForTest() {
	statusMu.Lock()
	status = Status{}
	statusEmit = nil
	statusMu.Unlock()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/dashboard/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/status.go internal/dashboard/status_test.go
git commit -m "feat: add dashboard status tracking package"
```

---

### Task 2: Dashboard upload counters

**Files:**
- Create: `internal/dashboard/counters.go`
- Test: `internal/dashboard/counters_test.go`

**Interfaces:**
- Consumes: nothing from Task 1.
- Produces: `func IncrementCounter(topic string)`, `func GetUploadCounts() map[string]int64`, `func OnCountersChange(fn func(map[string]int64))`, package var `CoalesceInterval time.Duration` (overridable by tests) — used by Task 5 (dispatcher) and Task 7 (main.go wiring).

- [ ] **Step 1: Write the failing tests**

```go
// internal/dashboard/counters_test.go
package dashboard

import (
	"testing"
	"time"
)

func TestIncrementCounter_AccumulatesPerTopic(t *testing.T) {
	resetCountersForTest()

	IncrementCounter("marketorders.ingest")
	IncrementCounter("marketorders.ingest")
	IncrementCounter("goldprices.ingest")

	got := GetUploadCounts()
	if got["marketorders.ingest"] != 2 {
		t.Fatalf("marketorders.ingest = %d, want 2", got["marketorders.ingest"])
	}
	if got["goldprices.ingest"] != 1 {
		t.Fatalf("goldprices.ingest = %d, want 1", got["goldprices.ingest"])
	}
}

func TestOnCountersChange_CoalescesAndOnlyEmitsOnChange(t *testing.T) {
	resetCountersForTest()
	CoalesceInterval = 10 * time.Millisecond

	var mu sync.Mutex
	var snapshots []map[string]int64
	OnCountersChange(func(c map[string]int64) {
		mu.Lock()
		defer mu.Unlock()
		cp := make(map[string]int64, len(c))
		for k, v := range c {
			cp[k] = v
		}
		snapshots = append(snapshots, cp)
	})

	IncrementCounter("marketorders.ingest")
	IncrementCounter("marketorders.ingest")
	time.Sleep(60 * time.Millisecond) // several ticks with no further change

	mu.Lock()
	got := len(snapshots)
	last := snapshots[len(snapshots)-1]
	mu.Unlock()

	if got == 0 {
		t.Fatal("expected at least one emitted snapshot")
	}
	if got > 3 {
		t.Fatalf("expected coalescing to keep emit count low, got %d emits for one change", got)
	}
	if last["marketorders.ingest"] != 2 {
		t.Fatalf("last snapshot marketorders.ingest = %d, want 2", last["marketorders.ingest"])
	}
}
```

Add `"sync"` to the test file's imports alongside `"testing"` and `"time"`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/dashboard/... -run TestIncrementCounter -v`
Expected: FAIL — `IncrementCounter`, `GetUploadCounts`, `resetCountersForTest` undefined.

- [ ] **Step 3: Implement**

```go
// internal/dashboard/counters.go
package dashboard

import (
	"sync"
	"sync/atomic"
	"time"
)

// CoalesceInterval is how often counter changes are checked and, if
// present, emitted as a single snapshot. Overridable in tests.
var CoalesceInterval = 500 * time.Millisecond

var (
	countersMu   sync.Mutex
	counters     = map[string]*int64{}
	lastSnapshot map[string]int64
	countersEmit func(map[string]int64)
	tickerOnce   sync.Once
)

// IncrementCounter records one upload attempt for the given topic
// (e.g. "marketorders.ingest").
func IncrementCounter(topic string) {
	countersMu.Lock()
	c, ok := counters[topic]
	if !ok {
		c = new(int64)
		counters[topic] = c
	}
	countersMu.Unlock()

	atomic.AddInt64(c, 1)
}

// GetUploadCounts returns a snapshot of all counters by topic.
func GetUploadCounts() map[string]int64 {
	countersMu.Lock()
	defer countersMu.Unlock()
	return snapshotLocked()
}

func snapshotLocked() map[string]int64 {
	out := make(map[string]int64, len(counters))
	for topic, c := range counters {
		out[topic] = atomic.LoadInt64(c)
	}
	return out
}

// OnCountersChange registers the callback invoked with a coalesced
// snapshot at most every CoalesceInterval, only when something changed
// since the last emit. Starts the background ticker on first call.
func OnCountersChange(fn func(map[string]int64)) {
	countersMu.Lock()
	countersEmit = fn
	countersMu.Unlock()

	tickerOnce.Do(func() {
		go coalesceLoop()
	})
}

func coalesceLoop() {
	ticker := time.NewTicker(CoalesceInterval)
	defer ticker.Stop()

	for range ticker.C {
		countersMu.Lock()
		next := snapshotLocked()
		changed := !countsEqual(next, lastSnapshot)
		if changed {
			lastSnapshot = next
		}
		emit := countersEmit
		countersMu.Unlock()

		if changed && emit != nil {
			emit(next)
		}
	}
}

func countsEqual(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// resetCountersForTest clears all package state. Test-only.
func resetCountersForTest() {
	countersMu.Lock()
	counters = map[string]*int64{}
	lastSnapshot = nil
	countersEmit = nil
	countersMu.Unlock()
}
```

Note: `tickerOnce` is intentionally never reset by `resetCountersForTest` — the ticker goroutine, once started, keeps running for the lifetime of the process (and the test binary), continuously reading the latest `countersEmit`/`counters` under the mutex. This mirrors production: there is exactly one dashboard per process.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/dashboard/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/counters.go internal/dashboard/counters_test.go
git commit -m "feat: add coalesced dashboard upload counters"
```

---

### Task 3: Dashboard log tail (logrus hook + ring buffer)

**Files:**
- Create: `internal/dashboard/loghook.go`
- Test: `internal/dashboard/loghook_test.go`

**Interfaces:**
- Consumes: `github.com/sirupsen/logrus` (already a project dependency, used by `log/logger.go`).
- Produces: `type LogLine struct { Level, Message string }`, `func NewLogHook() *LogHook` (where `*LogHook` implements `logrus.Hook`), `func GetRecentLogs() []LogLine`, `func OnLogLine(fn func(LogLine))` — used by Task 7 (`log.AddHook(dashboard.NewLogHook())`) and Task 4 (service).

- [ ] **Step 1: Write the failing tests**

```go
// internal/dashboard/loghook_test.go
package dashboard

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogHook_FireAppendsAndCaps(t *testing.T) {
	resetLogHookForTest()
	hook := NewLogHook()

	for i := 0; i < maxLogLines+10; i++ {
		if err := hook.Fire(&logrus.Entry{Level: logrus.InfoLevel, Message: "line"}); err != nil {
			t.Fatalf("Fire returned error: %v", err)
		}
	}

	got := GetRecentLogs()
	if len(got) != maxLogLines {
		t.Fatalf("len(GetRecentLogs()) = %d, want %d", len(got), maxLogLines)
	}
}

func TestLogHook_FireNotifiesSubscriber(t *testing.T) {
	resetLogHookForTest()
	hook := NewLogHook()

	var got []LogLine
	OnLogLine(func(l LogLine) { got = append(got, l) })

	_ = hook.Fire(&logrus.Entry{Level: logrus.WarnLevel, Message: "careful"})

	if len(got) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(got))
	}
	if got[0].Level != "warning" || got[0].Message != "careful" {
		t.Fatalf("unexpected log line: %+v", got[0])
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/dashboard/... -run TestLogHook -v`
Expected: FAIL — `NewLogHook`, `maxLogLines`, `GetRecentLogs`, `OnLogLine`, `resetLogHookForTest` undefined.

- [ ] **Step 3: Implement**

```go
// internal/dashboard/loghook.go
package dashboard

import (
	"sync"

	"github.com/sirupsen/logrus"
)

const maxLogLines = 500

// LogLine is one captured log entry, formatted for display.
type LogLine struct {
	Level   string
	Message string
}

var (
	logMu    sync.Mutex
	logLines []LogLine
	logEmit  func(LogLine)
)

// LogHook is a logrus.Hook that keeps the most recent maxLogLines log
// lines in memory and, if a subscriber is registered, forwards each new
// line as it arrives.
type LogHook struct{}

// NewLogHook returns a LogHook ready to be registered with log.AddHook.
func NewLogHook() *LogHook {
	return &LogHook{}
}

// Levels implements logrus.Hook: fire for every level.
func (h *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire implements logrus.Hook.
func (h *LogHook) Fire(entry *logrus.Entry) error {
	line := LogLine{Level: entry.Level.String(), Message: entry.Message}

	logMu.Lock()
	logLines = append(logLines, line)
	if len(logLines) > maxLogLines {
		logLines = logLines[len(logLines)-maxLogLines:]
	}
	emit := logEmit
	logMu.Unlock()

	if emit != nil {
		emit(line)
	}
	return nil
}

// OnLogLine registers the callback invoked once per new log line.
func OnLogLine(fn func(LogLine)) {
	logMu.Lock()
	logEmit = fn
	logMu.Unlock()
}

// GetRecentLogs returns a copy of the currently buffered log lines,
// oldest first.
func GetRecentLogs() []LogLine {
	logMu.Lock()
	defer logMu.Unlock()
	out := make([]LogLine, len(logLines))
	copy(out, logLines)
	return out
}

// resetLogHookForTest clears all package state. Test-only.
func resetLogHookForTest() {
	logMu.Lock()
	logLines = nil
	logEmit = nil
	logMu.Unlock()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/dashboard/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/loghook.go internal/dashboard/loghook_test.go
git commit -m "feat: add dashboard log tail ring buffer"
```

---

### Task 4: Dashboard bound service

**Files:**
- Create: `internal/dashboard/service.go`
- Test: `internal/dashboard/service_test.go`

**Interfaces:**
- Consumes: `GetStatus() Status` (Task 1), `GetUploadCounts() map[string]int64` (Task 2), `GetRecentLogs() []LogLine` (Task 3).
- Produces: `type DashboardService struct{}` with methods `GetStatus() Status`, `GetUploadCounts() map[string]int64`, `GetRecentLogs() []LogLine` — registered with `application.NewService` in Task 7, called from the frontend for backfill on window show (Task 9).

- [ ] **Step 1: Write the failing test**

```go
// internal/dashboard/service_test.go
package dashboard

import "testing"

func TestDashboardService_DelegatesToPackageState(t *testing.T) {
	resetStatusForTest()
	resetCountersForTest()
	resetLogHookForTest()

	SetVersionInfo("9.9.9", "")
	IncrementCounter("marketorders.ingest")
	hook := NewLogHook()
	_ = hook

	svc := &DashboardService{}

	if got := svc.GetStatus(); got.Version != "9.9.9" {
		t.Fatalf("GetStatus().Version = %q, want %q", got.Version, "9.9.9")
	}
	if got := svc.GetUploadCounts(); got["marketorders.ingest"] != 1 {
		t.Fatalf("GetUploadCounts()[marketorders.ingest] = %d, want 1", got["marketorders.ingest"])
	}
	if got := svc.GetRecentLogs(); len(got) != 0 {
		t.Fatalf("GetRecentLogs() = %+v, want empty (no Fire calls in this test)", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/dashboard/... -run TestDashboardService -v`
Expected: FAIL — `DashboardService` undefined.

- [ ] **Step 3: Implement**

```go
// internal/dashboard/service.go
package dashboard

// DashboardService is bound to the Wails frontend (see application.NewService
// in albiondata-client.go). Its methods back-fill the dashboard window's
// state when it's shown; ongoing updates arrive via the "status:changed",
// "counters:snapshot", and "log:line" events instead.
type DashboardService struct{}

// GetStatus returns the current status snapshot.
func (s *DashboardService) GetStatus() Status {
	return GetStatus()
}

// GetUploadCounts returns the current upload counters by topic.
func (s *DashboardService) GetUploadCounts() map[string]int64 {
	return GetUploadCounts()
}

// GetRecentLogs returns the currently buffered recent log lines.
func (s *DashboardService) GetRecentLogs() []LogLine {
	return GetRecentLogs()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/dashboard/... -v`
Expected: PASS (all `internal/dashboard` tests from Tasks 1-4)

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/service.go internal/dashboard/service_test.go
git commit -m "feat: add DashboardService for Wails frontend binding"
```

---

### Task 5: Wire dashboard state into the client package

**Files:**
- Modify: `client/listener.go:154-157`
- Modify: `client/albion_watcher.go` (`run` and `closeWatcher` methods)
- Modify: `client/dispatcher.go:55` (`sendMsgToPublicUploaders`)
- Test: `client/dispatcher_test.go`

**Interfaces:**
- Consumes: `dashboard.SetServer(serverID int, ingestBaseURL string)`, `dashboard.SetCaptureRunning(running bool)`, `dashboard.IncrementCounter(topic string)` (all from Task 1/2).

- [ ] **Step 1: Write the failing test**

```go
// client/dispatcher_test.go
package client

import (
	"testing"

	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/lib"
)

func TestSendMsgToPublicUploaders_IncrementsCounter(t *testing.T) {
	// No configured ingest targets: createUploaders returns nothing for
	// both public and private, so this exercises the counter increment
	// without making any network calls.
	ConfigGlobal.PublicIngestBaseUrls = ""
	ConfigGlobal.PrivateIngestBaseUrls = ""

	before := dashboard.GetUploadCounts()["marketorders.ingest"]

	sendMsgToPublicUploaders(struct{}{}, lib.NatsMarketOrdersIngest, &albionState{}, "test-id")

	after := dashboard.GetUploadCounts()["marketorders.ingest"]
	if after != before+1 {
		t.Fatalf("marketorders.ingest counter = %d, want %d", after, before+1)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./client/... -run TestSendMsgToPublicUploaders_IncrementsCounter -v`
Expected: FAIL — counter not incremented (0 != 1), since `sendMsgToPublicUploaders` doesn't call `dashboard.IncrementCounter` yet.

- [ ] **Step 3: Implement**

In `client/dispatcher.go`, add the import and the increment at the top of `sendMsgToPublicUploaders`:

```go
import (
	"encoding/json"
	"net/http"

	"strings"

	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/lib"
	"github.com/ao-data/albiondata-client/log"
)
```

```go
func sendMsgToPublicUploaders(upload interface{}, topic string, state *albionState, identifier string) {
	dashboard.IncrementCounter(topic)

	data, err := json.Marshal(upload)
	// ... rest unchanged
```

In `client/listener.go`, after line 157 (`log.Tracef("Using AODataIngestBaseURL: %s", ...)`), add:

```go
	dashboard.SetServer(l.router.albionstate.AODataServerID, l.router.albionstate.AODataIngestBaseURL)
```

and add the import:

```go
	"github.com/ao-data/albiondata-client/internal/dashboard"
```

In `client/albion_watcher.go`, add the import and two calls:

```go
import (
	"time"

	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/log"
)
```

```go
func (apw *albionProcessWatcher) run() error {
	log.Print("Watching Albion")
	physicalInterfaces, err := getAllPhysicalInterface()
	if err != nil {
		return err
	}
	apw.devices = physicalInterfaces
	log.Debugf("Will listen to these devices: %v", apw.devices)
	go apw.r.run()
	dashboard.SetCaptureRunning(true)

	for {
		// ... unchanged
```

```go
func (apw *albionProcessWatcher) closeWatcher() {
	dashboard.SetCaptureRunning(false)
	log.Print("Albion watcher closed")
	// ... unchanged
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./client/... ./internal/dashboard/... -v`
Expected: PASS, including all pre-existing `client` package tests (no regressions).

- [ ] **Step 5: Commit**

```bash
git add client/listener.go client/albion_watcher.go client/dispatcher.go client/dispatcher_test.go
git commit -m "feat: report server, capture, and upload state to the dashboard"
```

---

### Task 6: Cross-platform tray icon

**Files:**
- Create: `icon/tray.go`

**Interfaces:**
- Produces: `icon.TrayPNG []byte` — a PNG-encoded icon usable on all platforms, consumed by Task 7's tray setup (Wails' `SystemTray.SetIcon`/`SetTemplateIcon` expect PNG bytes, unlike the existing Windows `.ico`-based `icon.Data`).

- [ ] **Step 1: Implement**

```go
// icon/tray.go
package icon

import _ "embed"

// TrayPNG is the application icon in PNG form, used for the Wails v3
// system tray on all platforms (darwin/windows/linux). It is separate
// from the platform-specific Data variables (icondarwin.go, iconwin.go),
// which remain in their original .ico/.png forms for other uses (e.g.
// the Windows executable icon via go-winres).
//
//go:embed albiondata-client.png
var TrayPNG []byte
```

- [ ] **Step 2: Verify it builds**

Run: `go build ./icon/...`
Expected: success (no test needed — this is a data embed, exercised end-to-end in Task 7's manual verification).

- [ ] **Step 3: Commit**

```bash
git add icon/tray.go
git commit -m "feat: add cross-platform PNG tray icon"
```

---

### Task 7: Rewrite the entrypoint on Wails v3

**Files:**
- Delete: `systray/systray_darwin.go`, `systray/systray_win.go`, `systray/systray_others.go`
- Modify: `albiondata-client.go` (full rewrite)
- Modify: `go.mod`, `go.sum`, `vendor/` (via `go mod tidy && go mod vendor`)

**Interfaces:**
- Consumes: `dashboard.DashboardService` (Task 4), `dashboard.OnStatusChange`/`OnCountersChange`/`OnLogLine` (Tasks 1-3), `dashboard.NewLogHook` (Task 3), `icon.TrayPNG` (Task 6).
- Produces: `//go:embed all:frontend/dist` asset bundle that Task 8's frontend build populates, and the `"status:changed"` / `"counters:snapshot"` / `"log:line"` Wails events that Task 9's frontend subscribes to.

- [ ] **Step 1: Install the Wails v3 CLI and add the dependency**

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
go get github.com/wailsapp/wails/v3@latest
```

- [ ] **Step 2: Delete the old systray package**

```bash
rm systray/systray_darwin.go systray/systray_win.go systray/systray_others.go
rmdir systray
```

- [ ] **Step 3: Rewrite `albiondata-client.go`**

```go
package main

import (
	"embed"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/icon"
	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/log"

	"github.com/ao-data/go-githubupdate/updater"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

var version string

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	client.ConfigGlobal.SetupFlags()
	application.RegisterEvent[dashboard.Status]("status:changed")
	application.RegisterEvent[map[string]int64]("counters:snapshot")
	application.RegisterEvent[dashboard.LogLine]("log:line")
}

func main() {
	if client.ConfigGlobal.PrintVersion {
		log.Infof("Albion Data Client, version: %s", version)
		return
	}

	log.AddHook(dashboard.NewLogHook())

	startUpdater()

	// Wails owns the main thread on every platform, so packet capture
	// always runs in its own goroutine now (previously this was
	// conditional on darwin because only macOS's systray required the
	// main thread).
	go runClient()

	runDashboardApp()
}

func runClient() {
	c := client.NewClient(version)
	err := c.Run()
	if err != nil {
		log.Error(err)
		log.Error("The program encountered an error. Press any key to close this window.")
		var b = make([]byte, 1)
		_, _ = os.Stdin.Read(b)
	}
}

func runDashboardApp() {
	app := application.New(application.Options{
		Name:        "Albion Data Client",
		Description: "Live status dashboard for the Albion Data Client",
		Services: []application.Service{
			application.NewService(&dashboard.DashboardService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	dashboardWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Albion Data Client",
		Name:   "dashboard",
		Width:  900,
		Height: 600,
		Hidden: true,
		URL:    "/",
	})

	// Closing the window just hides it - the app (and packet capture)
	// keeps running, matching the old tray app's behavior.
	dashboardWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		dashboardWindow.Hide()
		e.Cancel()
	})

	dashboard.OnStatusChange(func(s dashboard.Status) {
		app.Event.Emit("status:changed", s)
	})
	dashboard.OnCountersChange(func(c map[string]int64) {
		app.Event.Emit("counters:snapshot", c)
	})
	dashboard.OnLogLine(func(l dashboard.LogLine) {
		app.Event.Emit("log:line", l)
	})

	setupTray(app, dashboardWindow)

	if err := app.Run(); err != nil {
		log.Error(err)
	}
}

func setupTray(app *application.App, dashboardWindow *application.WebviewWindow) {
	tray := app.SystemTray.New()
	tray.SetTooltip("Albion Data Client")

	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(icon.TrayPNG)
	} else {
		tray.SetIcon(icon.TrayPNG)
	}

	menu := app.NewMenu()
	menu.Add("Open Dashboard").OnClick(func(ctx *application.Context) {
		dashboardWindow.Show()
	})
	menu.Add("Open Log File").OnClick(func(ctx *application.Context) {
		openLogFile()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray.SetMenu(menu)
}

func openLogFile() {
	path := client.GetLogFilePath()
	if _, err := os.Stat(path); err != nil {
		log.Info("No log file found yet.")
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to open log file: %v", err)
	}
}

func startUpdater() {
	dashboard.SetVersionInfo(version, "")

	if version != "" && !strings.Contains(version, "dev") {
		u := updater.NewUpdater(
			version,
			client.ConfigGlobal.UpdateGithubOwner,
			client.ConfigGlobal.UpdateGithubRepo,
			"update-",
		)

		go func() {
			for {
				if tryUpdate(u) {
					restartProcess()
					return // This line won't be reached if restart succeeds, but included for clarity
				}
				// Wait 1 hour before checking again
				time.Sleep(time.Hour)
			}
		}()
	}
}

// tryUpdate attempts to check and apply an update with retry logic.
// Returns true if an update was successfully applied.
func tryUpdate(u *updater.Updater) bool {
	maxTries := 2
	for i := 0; i < maxTries; i++ {
		if available, checkErr := u.CheckUpdateAvailable(); checkErr == nil {
			dashboard.SetVersionInfo(version, available)
		}

		updated, err := u.BackgroundUpdater()
		if err != nil {
			log.Error(err.Error())
			if i < maxTries-1 {
				log.Info("Will try again in 60 seconds. You may need to run the client as Administrator.")
				time.Sleep(time.Second * 60)
			}
			continue
		}
		if updated {
			return true
		}
		// No update available, no need to retry
		return false
	}
	return false
}

// restartProcess replaces the current process with the updated version.
// On Unix systems (macOS/Linux), it uses syscall.Exec to seamlessly take over the terminal.
// On Windows, it starts a new process and exits since exec-style replacement isn't supported.
func restartProcess() {
	execPath, err := os.Executable()
	if err != nil {
		log.Errorf("Failed to get executable path for restart: %v", err)
		return
	}

	log.Info("Restarting with updated version...")

	if runtime.GOOS == "windows" {
		cmd := exec.Command(execPath, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Start()
		if err != nil {
			log.Errorf("Failed to start new process: %v", err)
			return
		}

		log.Info("New process started, exiting current process.")
		os.Exit(0)
	} else {
		err = syscall.Exec(execPath, os.Args, os.Environ())
		if err != nil {
			log.Errorf("Failed to exec new process: %v", err)
			cmd := exec.Command(execPath, os.Args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if startErr := cmd.Start(); startErr != nil {
				log.Errorf("Fallback process start also failed: %v", startErr)
				return
			}
			os.Exit(0)
		}
	}
}
```

This won't compile yet — `//go:embed all:frontend/dist` requires `frontend/dist/` to exist with at least one file, which Task 8 creates. Confirm the expected failure now so Task 8's success is unambiguous:

Run: `go build albiondata-client.go`
Expected: FAIL with `pattern all:frontend/dist: no matching files found` (or similar) — this is expected at this point in the plan.

- [ ] **Step 4: Update go.mod/go.sum/vendor**

```bash
go mod tidy
go mod vendor
```

Confirm `getlantern/systray` and `gonutz/w32` are gone and `wailsapp/wails/v3` is present:

```bash
grep -c "getlantern/systray\|gonutz/w32" go.mod   # expect: 0
grep -c "wailsapp/wails/v3" go.mod                # expect: >0
```

- [ ] **Step 5: Commit**

```bash
git add -A albiondata-client.go go.mod go.sum vendor systray
git commit -m "feat: replace systray with Wails v3 app bootstrap and tray"
```

(This intentionally leaves the build broken — `frontend/dist` doesn't exist yet. Task 8 fixes it. If your workflow requires every commit to build, squash Tasks 7-8 into one commit instead.)

---

### Task 8: Scaffold the Svelte + Vite frontend

**Files:**
- Delete: existing untracked `frontend/dist/`, `frontend/node_modules/` (stale leftovers from an earlier, abandoned GUI attempt)
- Create: `frontend/package.json`, `frontend/vite.config.js`, `frontend/index.html`, `frontend/src/main.js`, `frontend/.gitignore`

**Interfaces:**
- Produces: a Vite build that outputs to `frontend/dist/`, which Task 7's `//go:embed all:frontend/dist` picks up, and installs `@wailsio/runtime` for Task 9's event/service wiring.

- [ ] **Step 1: Clear stale leftovers and scaffold with Vite**

```bash
rm -rf frontend
npm create vite@latest frontend -- --template svelte
cd frontend
npm install
npm install @wailsio/runtime
cd ..
```

- [ ] **Step 2: Replace `frontend/index.html`**

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Albion Data Client</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.js"></script>
  </body>
</html>
```

- [ ] **Step 3: Confirm a production build produces `frontend/dist/`**

```bash
cd frontend && npm run build && cd ..
ls frontend/dist/index.html
```

Expected: the file exists.

- [ ] **Step 4: Confirm the Go build now succeeds**

```bash
go build albiondata-client.go
```

Expected: success (this was the failure left over from Task 7, now resolved).

- [ ] **Step 5: Add a `.gitignore` for frontend build/dependency output**

```
# frontend/.gitignore
node_modules
dist
```

- [ ] **Step 6: Commit**

```bash
rm -f albiondata-client   # don't commit the local build output
git add frontend
git commit -m "feat: scaffold Vite + Svelte frontend"
```

---

### Task 9: Dashboard UI (Svelte components)

**Files:**
- Create: `frontend/src/App.svelte`, `frontend/src/lib/StatusHeader.svelte`, `frontend/src/lib/CountersPanel.svelte`, `frontend/src/lib/LogTail.svelte`
- Modify: `frontend/src/main.js`
- Delete: default Vite scaffold files not used (`frontend/src/assets/`, `frontend/src/app.css` counterpart if unused — keep only what's referenced)

**Interfaces:**
- Consumes: Wails events `"status:changed"`, `"counters:snapshot"`, `"log:line"` (Task 7); bound service methods `GetStatus()`, `GetUploadCounts()`, `GetRecentLogs()` on `DashboardService` (Task 4), reached via generated bindings.

- [ ] **Step 1: Generate the Go bindings**

```bash
wails3 generate bindings ./internal/dashboard/...
```

Expected output directory: `frontend/bindings/github.com/ao-data/albiondata-client/internal/dashboard/`. List it and confirm the exact generated filename/export for `DashboardService`:

```bash
find frontend/bindings -iname "*dashboardservice*"
```

Use the path this prints in the import statements below — if it differs from `../bindings/github.com/ao-data/albiondata-client/internal/dashboard/dashboardservice` (e.g. different casing), adjust Step 4's import accordingly.

- [ ] **Step 2: `frontend/src/main.js`**

```js
import App from './App.svelte';
import { mount } from 'svelte';

const app = mount(App, {
  target: document.getElementById('app'),
});

export default app;
```

(If the Vite Svelte template scaffolded Svelte 4 instead of Svelte 5, use the Svelte 4 mounting form instead: `new App({ target: document.getElementById('app') })`. Check `frontend/package.json`'s `svelte` dependency version to decide.)

- [ ] **Step 3: `frontend/src/lib/StatusHeader.svelte`**

```svelte
<script>
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/dashboardservice';

  let status = {
    Version: '',
    UpdateAvailable: '',
    CaptureRunning: false,
    ServerID: 0,
    IngestBaseURL: '',
  };

  DashboardService.GetStatus().then((s) => (status = s));

  const unlisten = Events.On('status:changed', (evt) => {
    status = evt.data;
  });

  onDestroy(unlisten);

  const serverNames = { 0: 'Unknown', 1: 'West', 2: 'East', 3: 'Europe' };
</script>

<header class="status-header">
  <span class="badge" class:running={status.CaptureRunning}>
    {status.CaptureRunning ? 'Capturing' : 'Idle'}
  </span>
  <span>Server: {serverNames[status.ServerID] ?? status.ServerID}</span>
  <span>Version: {status.Version || 'dev'}</span>
  {#if status.UpdateAvailable}
    <span class="badge update">Update available: {status.UpdateAvailable}</span>
  {/if}
</header>

<style>
  .status-header {
    display: flex;
    gap: 1rem;
    align-items: center;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #333;
  }
  .badge {
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    background: #555;
  }
  .badge.running {
    background: #2e7d32;
  }
  .badge.update {
    background: #1565c0;
  }
</style>
```

- [ ] **Step 4: `frontend/src/lib/CountersPanel.svelte`**

```svelte
<script>
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/dashboardservice';

  const labels = {
    'marketorders.ingest': 'Market Orders',
    'goldprices.ingest': 'Gold Prices',
    'markethistories.ingest': 'Market Histories',
    'festivities.ingest': 'Festivities',
    'mapdata.ingest': 'Map Data',
    'banditevent.ingest': 'Bandit Events',
  };

  let counts = {};

  DashboardService.GetUploadCounts().then((c) => (counts = c));

  const unlisten = Events.On('counters:snapshot', (evt) => {
    counts = evt.data;
  });

  onDestroy(unlisten);
</script>

<section class="counters">
  {#each Object.entries(labels) as [topic, label]}
    <div class="counter">
      <span class="label">{label}</span>
      <span class="value">{counts[topic] ?? 0}</span>
    </div>
  {/each}
</section>

<style>
  .counters {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    padding: 1rem;
  }
  .counter {
    display: flex;
    flex-direction: column;
    min-width: 8rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid #333;
    border-radius: 6px;
  }
  .label {
    font-size: 0.8rem;
    opacity: 0.7;
  }
  .value {
    font-size: 1.5rem;
    font-weight: 600;
  }
</style>
```

- [ ] **Step 5: `frontend/src/lib/LogTail.svelte`**

```svelte
<script>
  import { onDestroy, tick } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/dashboardservice';

  let lines = [];
  let container;
  let stickToBottom = true;

  DashboardService.GetRecentLogs().then((initial) => {
    lines = initial;
    scrollToBottom();
  });

  const unlisten = Events.On('log:line', (evt) => {
    lines = [...lines, evt.data];
    scrollToBottom();
  });

  onDestroy(unlisten);

  async function scrollToBottom() {
    if (!stickToBottom) return;
    await tick();
    if (container) container.scrollTop = container.scrollHeight;
  }

  function handleScroll() {
    if (!container) return;
    const atBottom =
      container.scrollHeight - container.scrollTop - container.clientHeight < 20;
    stickToBottom = atBottom;
  }
</script>

<div class="log-tail" bind:this={container} on:scroll={handleScroll}>
  {#each lines as line}
    <div class="line level-{line.Level}">
      <span class="level">{line.Level}</span>
      <span class="message">{line.Message}</span>
    </div>
  {/each}
</div>

<style>
  .log-tail {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem 1rem;
    font-family: monospace;
    font-size: 0.85rem;
    background: #111;
  }
  .line {
    display: flex;
    gap: 0.5rem;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .level {
    text-transform: uppercase;
    opacity: 0.6;
    width: 5rem;
    flex-shrink: 0;
  }
  .level-error .message,
  .level-fatal .message,
  .level-panic .message {
    color: #ef5350;
  }
  .level-warning .message {
    color: #ffb74d;
  }
</style>
```

- [ ] **Step 6: `frontend/src/App.svelte`**

```svelte
<script>
  import StatusHeader from './lib/StatusHeader.svelte';
  import CountersPanel from './lib/CountersPanel.svelte';
  import LogTail from './lib/LogTail.svelte';
</script>

<main>
  <StatusHeader />
  <CountersPanel />
  <LogTail />
</main>

<style>
  :global(html, body) {
    margin: 0;
    height: 100%;
    background: #1a1a1a;
    color: #e0e0e0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  }
  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }
</style>
```

- [ ] **Step 7: Build and verify**

```bash
cd frontend && npm run build && cd ..
go build albiondata-client.go
```

Expected: both succeed.

- [ ] **Step 8: Commit**

```bash
rm -f albiondata-client
git add frontend
git commit -m "feat: add dashboard UI (status header, counters, log tail)"
```

---

### Task 10: Update build scripts for the frontend build step

**Files:**
- Modify: `scripts/build-darwin.sh`, `scripts/build-linux.sh`, `scripts/build-windows.sh`, `scripts/run.sh`
- Modify: `Makefile`

**Interfaces:**
- Consumes: `frontend/package.json`'s `build` script (Task 8/9).

- [ ] **Step 1: Add a frontend build step before each `go build`**

In `scripts/build-darwin.sh`, `scripts/build-linux.sh`, and `scripts/build-windows.sh`, insert immediately before the existing `go build` line:

```bash
(cd frontend && npm ci && npm run build)
```

In `scripts/run.sh`, insert the same line before the `go run` line:

```bash
#!/usr/bin/env bash

set -e

(cd frontend && npm ci && npm run build)
go run -ldflags="-w -s -X main.version=dev" albiondata-client.go
```

- [ ] **Step 2: Add a `frontend` target to the `Makefile`**

```makefile
frontend:
	cd frontend && npm ci && npm run build
```

- [ ] **Step 3: Verify `scripts/run.sh` works end to end**

```bash
./scripts/run.sh -h
```

Expected: the client's `-h` flag output prints (confirms the frontend build + `go run` both succeeded and the binary runs).

- [ ] **Step 4: Commit**

```bash
git add scripts/build-darwin.sh scripts/build-linux.sh scripts/build-windows.sh scripts/run.sh Makefile
git commit -m "build: build the frontend before the Go binary"
```

---

### Task 11: Manual verification

No new files — this task exercises the full feature per the project's verification standard (UI changes must be manually tested, not just type-checked).

- [ ] **Step 1: Run the client in dev mode**

```bash
./scripts/run.sh
```

- [ ] **Step 2: Verify the tray icon appears** with "Open Dashboard", "Open Log File", and "Quit" items (on whichever of macOS/Windows/Linux you're testing on; note in your report if Linux has no notification-area extension, per the documented limitation).

- [ ] **Step 3: Click "Open Dashboard"** and verify the window appears showing:
  - The status header (capture state, server, version)
  - The counters panel (all zero initially, since no Albion Online traffic has been captured yet)
  - An empty log tail that then fills with startup log lines

- [ ] **Step 4: Close the dashboard window** (titlebar close button) and verify the app keeps running (check the process, or click "Open Dashboard" again and confirm it reappears with continuity — e.g. log lines logged while it was closed are present).

- [ ] **Step 5: If you have a way to generate Albion Online traffic** (a running game client on this machine), verify:
  - The status header's capture badge flips to "Capturing" and the server field populates
  - The counters panel increments as market data is captured
  - The log tail streams new lines in real time

  If you don't have game traffic available, note this in your report as unverified rather than claiming it works.

- [ ] **Step 6: Click "Open Log File"** and verify it opens the log file in the OS default viewer/editor.

- [ ] **Step 7: Click "Quit"** and verify the process exits.

- [ ] **Step 8: Run the full test suite one more time**

```bash
go test ./...
```

Expected: PASS, no regressions.

Report the results of Steps 2-7 (what worked, what didn't, what was unverifiable) rather than a blanket "done."
