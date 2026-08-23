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
