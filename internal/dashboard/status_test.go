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

func TestSetCustomPublicIngest_EmitsOnChange(t *testing.T) {
	resetStatusForTest()

	var got []Status
	OnStatusChange(func(s Status) { got = append(got, s) })

	SetCustomPublicIngest(true)
	SetCustomPublicIngest(true) // no change, no emit
	SetCustomPublicIngest(false)

	if len(got) != 2 {
		t.Fatalf("expected 2 emits, got %d: %+v", len(got), got)
	}
	if got[0].CustomPublicIngest != true || got[1].CustomPublicIngest != false {
		t.Fatalf("unexpected emitted statuses: %+v", got)
	}
}

func TestSetCaptureError_EmitsOnChange(t *testing.T) {
	resetStatusForTest()

	var got []Status
	OnStatusChange(func(s Status) { got = append(got, s) })

	SetCaptureError(true)
	SetCaptureError(true) // no change, no emit
	SetCaptureError(false)

	if len(got) != 2 {
		t.Fatalf("expected 2 emits, got %d: %+v", len(got), got)
	}
	if got[0].CaptureError != true || got[1].CaptureError != false {
		t.Fatalf("unexpected emitted statuses: %+v", got)
	}
}

func TestSetEncryptionStatus_EmitsOnChange(t *testing.T) {
	resetStatusForTest()

	var got []Status
	OnStatusChange(func(s Status) { got = append(got, s) })

	SetEncryptionStatus(EncryptionDetected)
	SetEncryptionStatus(EncryptionDetected) // no change, no emit
	SetEncryptionStatus(EncryptionClear)
	SetEncryptionStatus(EncryptionUnknown)

	if len(got) != 3 {
		t.Fatalf("expected 3 emits, got %d: %+v", len(got), got)
	}
	if got[0].EncryptionStatus != EncryptionDetected ||
		got[1].EncryptionStatus != EncryptionClear ||
		got[2].EncryptionStatus != EncryptionUnknown {
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
