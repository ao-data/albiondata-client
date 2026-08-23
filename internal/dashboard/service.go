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
