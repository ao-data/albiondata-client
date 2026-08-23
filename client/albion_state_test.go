package client

import (
	"testing"
	"time"
)

func TestShouldNotifyMarketDataEncrypted_WithinWindowAfterRequest(t *testing.T) {
	state := &albionState{}
	base := time.Now()

	state.LastMarketDataRequestAt = base

	if !state.ShouldNotifyMarketDataEncrypted(base.Add(1 * time.Second)) {
		t.Fatal("expected notification when encrypted command seen shortly after a market data request")
	}
}

func TestShouldNotifyMarketDataEncrypted_OutsideWindow(t *testing.T) {
	state := &albionState{}
	base := time.Now()

	state.LastMarketDataRequestAt = base

	if state.ShouldNotifyMarketDataEncrypted(base.Add(5 * time.Second)) {
		t.Fatal("did not expect notification when encrypted command seen well after the correlation window")
	}
}

func TestShouldNotifyMarketDataEncrypted_NoRecentRequest(t *testing.T) {
	state := &albionState{}

	if state.ShouldNotifyMarketDataEncrypted(time.Now()) {
		t.Fatal("did not expect notification when no market data request has been seen")
	}
}

func TestShouldNotifyMarketDataEncrypted_OnlyNotifiesOnce(t *testing.T) {
	state := &albionState{}
	base := time.Now()

	state.LastMarketDataRequestAt = base

	if !state.ShouldNotifyMarketDataEncrypted(base.Add(1 * time.Second)) {
		t.Fatal("expected first call to notify")
	}

	if state.ShouldNotifyMarketDataEncrypted(base.Add(1 * time.Second)) {
		t.Fatal("did not expect a second notification for the same encryption event")
	}
}

func TestRecordMarketDataRequest_RearmsNotificationGuard(t *testing.T) {
	state := &albionState{MarketDataEncryptionNotified: true}
	now := time.Now()

	state.RecordMarketDataRequest(now)

	if state.MarketDataEncryptionNotified {
		t.Fatal("expected recording a new market data request to re-arm the notification guard")
	}
	if !state.LastMarketDataRequestAt.Equal(now) {
		t.Fatalf("expected LastMarketDataRequestAt to be %v, got %v", now, state.LastMarketDataRequestAt)
	}
}

func TestShouldNotifyMarketDataEncrypted_NotifiesOnEachNewAttempt(t *testing.T) {
	state := &albionState{}
	t0 := time.Now()

	state.RecordMarketDataRequest(t0)

	if !state.ShouldNotifyMarketDataEncrypted(t0.Add(1 * time.Second)) {
		t.Fatal("expected notification on first attempt")
	}
	if state.ShouldNotifyMarketDataEncrypted(t0.Add(1 * time.Second)) {
		t.Fatal("did not expect a duplicate notification within the same attempt")
	}

	t1 := t0.Add(10 * time.Second)
	state.RecordMarketDataRequest(t1)

	if !state.ShouldNotifyMarketDataEncrypted(t1.Add(1 * time.Second)) {
		t.Fatal("expected notification again on a new market data attempt")
	}
}
