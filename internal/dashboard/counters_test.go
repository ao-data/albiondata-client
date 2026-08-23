// internal/dashboard/counters_test.go
package dashboard

import (
	"sync"
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
