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
	IncrementCounterBy(topic, 1)
}

// IncrementCounterBy adds delta to the counter for the given topic.
// Used when a single upload contains multiple records (e.g. a market
// orders pack with 50 orders in it), so the dashboard reflects the
// number of records actually sent rather than the number of upload
// packs.
func IncrementCounterBy(topic string, delta int64) {
	countersMu.Lock()
	c, ok := counters[topic]
	if !ok {
		c = new(int64)
		counters[topic] = c
	}
	countersMu.Unlock()

	atomic.AddInt64(c, delta)
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

// ResetCounters clears all upload counters back to zero and emits the
// resulting empty snapshot. Called when capture goes idle, since counts
// from a finished capture session aren't meaningful for whatever session
// comes next.
func ResetCounters() {
	countersMu.Lock()
	counters = map[string]*int64{}
	next := map[string]int64{}
	lastSnapshot = next
	emit := countersEmit
	countersMu.Unlock()

	if emit != nil {
		emit(next)
	}
}

// resetCountersForTest clears all package state. Test-only.
func resetCountersForTest() {
	countersMu.Lock()
	counters = map[string]*int64{}
	lastSnapshot = nil
	countersEmit = nil
	countersMu.Unlock()
}
