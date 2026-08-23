# Dashboard Capturing Status + Private Server Label Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The dashboard's status badge shows "Ready" (light blue) until real game traffic arrives, "Capturing" (green) while it's flowing, and reverts to "Ready" after 30s of silence; the Server field shows "Private" when the user has configured a custom (non-default) `-i` ingest URL.

**Architecture:** A new `internal/dashboard/activity.go` tracks a rolling "last activity" timestamp and a background staleness ticker, both following the exact patterns already established in `internal/dashboard/counters.go` (a `sync.Once`-started ticker goroutine, an exported overridable interval for tests). It's wired into `client/listener.go` at the single point every successfully-decoded message already funnels through (`dispatchOperation`), replacing the old watcher-alive signal in `client/albion_watcher.go`. A new `Status.PrivateIngest` field is computed once at startup in `albiondata-client.go`. The frontend changes are confined to `StatusHeader.svelte`.

**Tech Stack:** Go 1.24 (existing), Svelte 5 (existing).

## Global Constraints

- "Successfully decoded message" (the activity signal) means any Photon `OperationRequest`/`OperationResponse`/`EventData` that decoded without error — regardless of whether the underlying Photon command was transport-reliable or unreliable. Decode errors do not count as activity.
- The staleness threshold is 30 seconds of no activity.
- `PrivateIngest` is `true` whenever `client.ConfigGlobal.PublicIngestBaseUrls != "https+pow://albion-online-data.com"` (the exact default string for the `-i` flag, from `client/config.go`) — computed once at startup, never recomputed.
- `client/albion_watcher.go`'s existing `dashboard.SetCaptureRunning(true)` / `dashboard.SetCaptureRunning(false)` calls are removed — the activity heartbeat is the only thing that sets `CaptureRunning` going forward.
- No changes to actual packet capture/decoding logic or the upload/ingest pipeline — this plan only adds observability.

---

## File Structure

New:
- `internal/dashboard/activity.go` — activity heartbeat + staleness ticker
- `internal/dashboard/activity_test.go`

Modified:
- `internal/dashboard/status.go` — add `PrivateIngest` field to `Status`, add `SetPrivateIngest` setter
- `internal/dashboard/status_test.go` — test for `SetPrivateIngest`
- `client/listener.go` — call `dashboard.RecordActivity()` from `dispatchOperation`
- `client/albion_watcher.go` — remove the two `dashboard.SetCaptureRunning` calls
- `albiondata-client.go` — call `dashboard.SetPrivateIngest(...)` once at startup
- `frontend/src/lib/StatusHeader.svelte` — badge relabel/recolor, "Private" server label

---

### Task 1: Activity heartbeat in `internal/dashboard`

**Files:**
- Create: `internal/dashboard/activity.go`
- Test: `internal/dashboard/activity_test.go`

**Interfaces:**
- Consumes: `SetCaptureRunning(bool)`, `OnStatusChange(func(Status))`, `GetStatus() Status`, `resetStatusForTest()` (all from `status.go`, already exist).
- Produces: `func RecordActivity()`, exported vars `CaptureStaleAfter time.Duration` (default 30s) and `StaleCheckInterval time.Duration` (default 5s) — both used by Task 3's wiring and overridable in tests. `resetActivityForTest()` for test use (mirrors `resetStatusForTest`/`resetCountersForTest`).

- [ ] **Step 1: Write the failing tests**

```go
// internal/dashboard/activity_test.go
package dashboard

import (
	"sync"
	"testing"
	"time"
)

// Force the staleness ticker to a fast, fixed tick rate for the whole
// test binary. StaleCheckInterval is baked into time.NewTicker at the
// moment the ticker goroutine is first started (via sync.Once), so it
// must be set before any test's first RecordActivity() call - setting it
// per-test would be a no-op for every test after the first one in
// execution order. CaptureStaleAfter is read fresh on every tick, so it's
// safe to override per-test.
func init() {
	StaleCheckInterval = 5 * time.Millisecond
}

func TestRecordActivity_SetsCaptureRunningAndEmitsOnce(t *testing.T) {
	resetStatusForTest()
	resetActivityForTest()
	CaptureStaleAfter = time.Hour // keep this test's goroutine from going stale mid-test

	var mu sync.Mutex
	var got []Status
	OnStatusChange(func(s Status) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	})

	RecordActivity()
	RecordActivity()

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 emit for the false->true transition, got %d: %+v", len(got), got)
	}
	if !got[0].CaptureRunning {
		t.Fatalf("expected CaptureRunning=true, got %+v", got[0])
	}
}

func TestRecordActivity_GoesStaleAfterTimeout(t *testing.T) {
	resetStatusForTest()
	resetActivityForTest()
	CaptureStaleAfter = 20 * time.Millisecond

	var mu sync.Mutex
	var got []Status
	OnStatusChange(func(s Status) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	})

	RecordActivity()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(got) < 2 {
		t.Fatalf("expected at least 2 emits (true then false), got %d: %+v", len(got), got)
	}
	last := got[len(got)-1]
	if last.CaptureRunning {
		t.Fatalf("expected CaptureRunning=false after the staleness timeout, got %+v", last)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/dashboard/... -run TestRecordActivity -v`
Expected: FAIL — `RecordActivity`, `resetActivityForTest`, `StaleCheckInterval`, `CaptureStaleAfter` undefined (package doesn't exist yet).

- [ ] **Step 3: Implement**

```go
// internal/dashboard/activity.go
package dashboard

import (
	"sync"
	"time"
)

// CaptureStaleAfter is how long without a RecordActivity() call before
// CaptureRunning is reported false again. Overridable in tests.
var CaptureStaleAfter = 30 * time.Second

// StaleCheckInterval is how often the staleness ticker checks for
// inactivity. Overridable in tests - but only takes effect if set before
// the first RecordActivity() call in the process, since it's read once
// when the ticker goroutine starts (see RecordActivity).
var StaleCheckInterval = 5 * time.Second

var (
	activityMu   sync.Mutex
	lastActivity time.Time
	activityOnce sync.Once
)

// RecordActivity records that a decoded game message was just received,
// marking capture as running (if not already) and resetting the
// inactivity timer. Starts the background staleness ticker on first call.
func RecordActivity() {
	activityMu.Lock()
	lastActivity = time.Now()
	activityMu.Unlock()

	SetCaptureRunning(true)

	activityOnce.Do(func() {
		go staleCheckLoop()
	})
}

func staleCheckLoop() {
	ticker := time.NewTicker(StaleCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		activityMu.Lock()
		idle := time.Since(lastActivity)
		activityMu.Unlock()

		if idle >= CaptureStaleAfter {
			SetCaptureRunning(false)
		}
	}
}

// resetActivityForTest clears activity-tracking state. Test-only. Does
// not reset activityOnce - matching counters.go's tickerOnce precedent,
// the ticker goroutine keeps running for the test binary's lifetime once
// started, reading the (possibly test-overridden) package vars fresh on
// every tick.
func resetActivityForTest() {
	activityMu.Lock()
	lastActivity = time.Time{}
	activityMu.Unlock()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/dashboard/... -v`
Expected: PASS (all tests in the package, including the pre-existing ones from earlier plans)

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/activity.go internal/dashboard/activity_test.go
git commit -m "feat: add dashboard activity heartbeat with staleness timeout"
```

---

### Task 2: `PrivateIngest` field and setter

**Files:**
- Modify: `internal/dashboard/status.go`
- Modify: `internal/dashboard/status_test.go`

**Interfaces:**
- Produces: `Status.PrivateIngest bool` field, `func SetPrivateIngest(private bool)` — used by Task 4's wiring in `albiondata-client.go`.

- [ ] **Step 1: Write the failing test**

Add to `internal/dashboard/status_test.go`:

```go
func TestSetPrivateIngest_EmitsOnChange(t *testing.T) {
	resetStatusForTest()

	var got []Status
	OnStatusChange(func(s Status) { got = append(got, s) })

	SetPrivateIngest(true)
	SetPrivateIngest(true) // no change, no emit
	SetPrivateIngest(false)

	if len(got) != 2 {
		t.Fatalf("expected 2 emits, got %d: %+v", len(got), got)
	}
	if got[0].PrivateIngest != true || got[1].PrivateIngest != false {
		t.Fatalf("unexpected emitted statuses: %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/dashboard/... -run TestSetPrivateIngest_EmitsOnChange -v`
Expected: FAIL — `SetPrivateIngest` undefined, and the `Status` struct literal in the test won't compile against the field yet either.

- [ ] **Step 3: Implement**

In `internal/dashboard/status.go`, add the field to the `Status` struct:

```go
// Status is a point-in-time snapshot of the client's observable state,
// pushed to the dashboard frontend whenever it changes.
type Status struct {
	Version         string
	UpdateAvailable string
	CaptureRunning  bool
	ServerID        int
	IngestBaseURL   string
	PrivateIngest   bool
}
```

Add the setter (place it near `SetServer`, since both concern the ingest/server configuration):

```go
// SetPrivateIngest records whether the user has configured a custom
// (non-default) public ingest URL, meaning uploads go to a self-hosted
// server instead of the Albion Data Project's public endpoint.
func SetPrivateIngest(private bool) {
	next := GetStatus()
	next.PrivateIngest = private
	setStatus(next)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/dashboard/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/dashboard/status.go internal/dashboard/status_test.go
git commit -m "feat: add PrivateIngest field to dashboard status"
```

---

### Task 3: Wire the activity heartbeat into packet decoding, remove the old watcher-alive signal

**Files:**
- Modify: `client/listener.go`
- Modify: `client/albion_watcher.go`

**Interfaces:**
- Consumes: `dashboard.RecordActivity()` (Task 1).

- [ ] **Step 1: Add the activity hook to `dispatchOperation`**

In `client/listener.go`, find `dispatchOperation` (currently the last-called function in `onRequest`/`onResponse`/`onEvent` — this is the single point every successfully-decoded message already funnels through, so one hook here covers all three, rather than three separate call sites):

```go
func (l *listener) dispatchOperation(op operation, err error, params map[byte]interface{}) {
	if err != nil && !ConfigGlobal.DebugIgnoreDecodingErrors {
		log.Debugf("Error while decoding an event or operation: %v - params: %s", err, formatDebugPhotonParams(params))
		return
	}
	if op != nil {
		l.router.newOperation <- op
	}
}
```

Change to:

```go
func (l *listener) dispatchOperation(op operation, err error, params map[byte]interface{}) {
	if err != nil && !ConfigGlobal.DebugIgnoreDecodingErrors {
		log.Debugf("Error while decoding an event or operation: %v - params: %s", err, formatDebugPhotonParams(params))
		return
	}
	if err == nil {
		dashboard.RecordActivity()
	}
	if op != nil {
		l.router.newOperation <- op
	}
}
```

The `err == nil` check (rather than just relying on the early return above) matters because `ConfigGlobal.DebugIgnoreDecodingErrors` can let execution continue past the error check even when `err != nil` — activity should only be recorded for messages that actually decoded successfully, not ones that errored but were tolerated by that debug flag.

`client/listener.go` already imports `"github.com/ao-data/albiondata-client/internal/dashboard"` (used by the existing `dashboard.SetServer` call in `processPacket`), so no new import is needed.

- [ ] **Step 2: Remove the old watcher-alive signal**

In `client/albion_watcher.go`, remove both `dashboard.SetCaptureRunning` calls:

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
```

becomes:

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

	for {
```

and:

```go
func (apw *albionProcessWatcher) closeWatcher() {
	dashboard.SetCaptureRunning(false)
	log.Print("Albion watcher closed")
```

becomes:

```go
func (apw *albionProcessWatcher) closeWatcher() {
	log.Print("Albion watcher closed")
```

Since `dashboard.SetCaptureRunning` is no longer called anywhere in this file, remove the now-unused import:

```go
import (
	"time"

	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/log"
)
```

becomes:

```go
import (
	"time"

	"github.com/ao-data/albiondata-client/log"
)
```

- [ ] **Step 3: Run tests to verify no regressions**

Run: `go test ./client/... ./internal/dashboard/... -v`
Expected: PASS — all existing tests, plus Tasks 1-2's new tests, still pass. No new tests are added in this task (it's wiring existing, already-tested behavior into new call sites, plus a deletion) — behavior is exercised end-to-end in Task 5's manual verification.

- [ ] **Step 4: Verify the full build still succeeds**

```bash
go build ./client/... ./internal/... ./icon/...
```
Expected: success.

- [ ] **Step 5: Commit**

```bash
git add client/listener.go client/albion_watcher.go
git commit -m "feat: drive capture status from decoded message activity, not watcher liveness"
```

---

### Task 4: Compute `PrivateIngest` at startup

**Files:**
- Modify: `albiondata-client.go`

**Interfaces:**
- Consumes: `dashboard.SetPrivateIngest(bool)` (Task 2).

- [ ] **Step 1: Add the startup call**

In `albiondata-client.go`, find `startUpdater()`:

```go
func startUpdater() {
	dashboard.SetVersionInfo(version, "")

	if version != "" && !strings.Contains(version, "dev") {
```

Change to:

```go
func startUpdater() {
	dashboard.SetVersionInfo(version, "")
	dashboard.SetPrivateIngest(client.ConfigGlobal.PublicIngestBaseUrls != "https+pow://albion-online-data.com")

	if version != "" && !strings.Contains(version, "dev") {
```

`albiondata-client.go` already imports both `"github.com/ao-data/albiondata-client/client"` and `"github.com/ao-data/albiondata-client/internal/dashboard"`, so no new imports are needed. The literal string must match `client/config.go`'s `-i` flag default exactly (`https+pow://albion-online-data.com`) — this is intentionally not extracted into a shared constant in this plan, since `client/config.go`'s flag default is itself a literal in a `flag.StringVar` call, not an exported constant, and introducing one is out of scope for this plan (YAGNI: one call site each).

- [ ] **Step 2: Verify it builds**

```bash
go build ./client/... ./internal/... ./icon/...
go build albiondata-client.go
```
Expected: both succeed.

- [ ] **Step 3: Commit**

```bash
git add albiondata-client.go
git commit -m "feat: compute PrivateIngest status from the -i flag at startup"
```

---

### Task 5: Frontend badge + server label

**Files:**
- Modify: `frontend/src/lib/StatusHeader.svelte`

**Interfaces:**
- Consumes: `status.CaptureRunning`, `status.PrivateIngest`, `status.ServerID` (all already present in the `Status` model the frontend receives — Task 2 adds `PrivateIngest` to the Go struct, which flows through automatically via the existing `status:changed` event and `GetStatus()` binding; no bindings regeneration is needed beyond what Task 2's build already produces, but see Step 1).

- [ ] **Step 1: Regenerate Wails bindings**

Since Task 2 added a field to the Go `Status` struct, regenerate the JS model so the frontend's type information matches:

```bash
wails3 generate bindings ./...
```

Confirm `frontend/bindings/github.com/ao-data/albiondata-client/internal/dashboard/models.js` now includes `PrivateIngest` in the `Status` class (check with `grep -n "PrivateIngest" frontend/bindings/github.com/ao-data/albiondata-client/internal/dashboard/models.js` — expect at least one match). This is a generated file; if the diff also touches unrelated generated content (e.g. a regenerated timestamp comment), that's expected and fine to include in the commit.

- [ ] **Step 2: Update `StatusHeader.svelte`**

Current:

```svelte
<script>
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/index.js';

  let status = $state({
    Version: '',
    UpdateAvailable: '',
    CaptureRunning: false,
    ServerID: 0,
    IngestBaseURL: '',
  });

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

Replace with:

```svelte
<script>
  import { onDestroy } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { DashboardService } from '../../bindings/github.com/ao-data/albiondata-client/internal/dashboard/index.js';

  let status = $state({
    Version: '',
    UpdateAvailable: '',
    CaptureRunning: false,
    ServerID: 0,
    IngestBaseURL: '',
    PrivateIngest: false,
  });

  DashboardService.GetStatus().then((s) => (status = s));

  const unlisten = Events.On('status:changed', (evt) => {
    status = evt.data;
  });

  onDestroy(unlisten);

  const serverNames = { 0: 'Unknown', 1: 'West', 2: 'East', 3: 'Europe' };
  let serverLabel = $derived(
    status.PrivateIngest ? 'Private' : (serverNames[status.ServerID] ?? status.ServerID)
  );
</script>

<header class="status-header">
  <span class="badge" class:running={status.CaptureRunning}>
    {status.CaptureRunning ? 'Capturing' : 'Ready'}
  </span>
  <span>Server: {serverLabel}</span>
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
    background: #4fc3f7;
    color: #0d1b1e;
  }
  .badge.running {
    background: #2e7d32;
    color: inherit;
  }
  .badge.update {
    background: #1565c0;
    color: inherit;
  }
</style>
```

The badge's default (non-running) background is now `#4fc3f7` (light blue) instead of `#555` (gray), with a dark text color (`#0d1b1e`) for contrast against the light background — the `.running`/`.update` variants keep their existing darker backgrounds and revert to the header's default (light) text color via `color: inherit`.

- [ ] **Step 3: Build and verify**

```bash
cd frontend && npm run build && cd ..
go build albiondata-client.go
```
Expected: both succeed.

- [ ] **Step 4: Manual verification**

Run the built binary. Confirm:
- The badge shows "Ready" with a light blue background immediately at launch.
- If you have Albion Online traffic available (a running game client on this machine), confirm the badge flips to "Capturing" (green) within a moment of the game generating traffic, and reverts to "Ready" after roughly 30 seconds of no further traffic (e.g. after closing the game or going idle in a loading screen). If you don't have game traffic available, note this as unverified rather than claiming it works, per this project's standard verification practice.
- Launch with a custom `-i` value (e.g. `-i "nats://localhost:4222"`) and confirm the Server field shows "Private" instead of a region name. Launch normally (no `-i`) and confirm it still shows the detected region (or "Unknown" before any server is detected).

- [ ] **Step 5: Commit**

```bash
rm -f albiondata-client
git add frontend/src/lib/StatusHeader.svelte frontend/bindings
git commit -m "feat: show live capturing status and private-ingest server label"
```
