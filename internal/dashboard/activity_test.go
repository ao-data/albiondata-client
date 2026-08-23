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
	setCaptureStaleAfterForTest(time.Hour) // keep this test's goroutine from going stale mid-test
	t.Cleanup(func() { setCaptureStaleAfterForTest(time.Hour) })

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
	setCaptureStaleAfterForTest(20 * time.Millisecond)
	t.Cleanup(func() { setCaptureStaleAfterForTest(time.Hour) })

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

func TestRecordActivity_GoingStaleClearsCountersAndServer(t *testing.T) {
	resetStatusForTest()
	resetActivityForTest()
	resetCountersForTest()
	setCaptureStaleAfterForTest(20 * time.Millisecond)
	t.Cleanup(func() { setCaptureStaleAfterForTest(time.Hour) })

	IncrementCounter("marketorders.ingest")
	SetServer(1, "https://example.invalid")
	SetEncryptionStatus(EncryptionClear)

	RecordActivity()
	time.Sleep(100 * time.Millisecond)

	if GetStatus().CaptureRunning {
		t.Fatalf("expected capture to have gone stale")
	}
	if counts := GetUploadCounts(); len(counts) != 0 {
		t.Fatalf("expected counters cleared after going stale, got %+v", counts)
	}
	got := GetStatus()
	if got.ServerID != 0 || got.IngestBaseURL != "" {
		t.Fatalf("expected server cleared after going stale, got ServerID=%d IngestBaseURL=%q", got.ServerID, got.IngestBaseURL)
	}
	if got.EncryptionStatus != EncryptionUnknown {
		t.Fatalf("expected encryption status cleared after going stale, got %q", got.EncryptionStatus)
	}
}

func TestRecordActivity_StillRunningDoesNotClearCountersOrServer(t *testing.T) {
	resetStatusForTest()
	resetActivityForTest()
	resetCountersForTest()
	setCaptureStaleAfterForTest(time.Hour) // never goes stale during this test
	t.Cleanup(func() { setCaptureStaleAfterForTest(time.Hour) })

	IncrementCounter("marketorders.ingest")
	SetServer(1, "https://example.invalid")
	SetEncryptionStatus(EncryptionClear)

	RecordActivity()

	if counts := GetUploadCounts(); counts["marketorders.ingest"] != 1 {
		t.Fatalf("expected counter to survive while still running, got %+v", counts)
	}
	got := GetStatus()
	if got.ServerID != 1 {
		t.Fatalf("expected server to survive while still running, got ServerID=%d", got.ServerID)
	}
	if got.EncryptionStatus != EncryptionClear {
		t.Fatalf("expected encryption status to survive while still running, got %q", got.EncryptionStatus)
	}
}
