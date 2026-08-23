// internal/dashboard/activity.go
package dashboard

import (
	"sync"
	"time"
)

// StaleCheckInterval is how often the staleness ticker checks for
// inactivity. Overridable in tests - but only takes effect if set before
// the first RecordActivity() call in the process, since it's read once
// when the ticker goroutine starts (see RecordActivity).
var StaleCheckInterval = 5 * time.Second

var (
	activityMu        sync.Mutex
	lastActivity      time.Time
	activityOnce      sync.Once
	captureStaleAfter = 10 * time.Second
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
		staleAfter := captureStaleAfter
		activityMu.Unlock()

		if idle >= staleAfter {
			wasRunning := GetStatus().CaptureRunning
			SetCaptureRunning(false)
			if wasRunning {
				// Capture just went idle - counts, the detected server, and
				// the last observed encryption status are no longer
				// meaningful for whatever comes next.
				ResetCounters()
				SetServer(0, "")
				SetEncryptionStatus(EncryptionUnknown)
			}
		}
	}
}

// setCaptureStaleAfterForTest overrides the staleness threshold used by
// the background ticker. Test-only; safe for concurrent use with
// staleCheckLoop since both go through activityMu.
func setCaptureStaleAfterForTest(d time.Duration) {
	activityMu.Lock()
	captureStaleAfter = d
	activityMu.Unlock()
}

// resetActivityForTest clears activity-tracking state and restores the
// default staleness threshold. Test-only. Does not reset activityOnce -
// matching counters.go's tickerOnce precedent, the ticker goroutine keeps
// running for the test binary's lifetime once started, reading the
// (possibly test-overridden) package vars fresh on every tick.
func resetActivityForTest() {
	activityMu.Lock()
	lastActivity = time.Time{}
	captureStaleAfter = 10 * time.Second
	activityMu.Unlock()
}
