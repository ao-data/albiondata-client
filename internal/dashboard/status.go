// internal/dashboard/status.go
package dashboard

import "sync"

// Status is a point-in-time snapshot of the client's observable state,
// pushed to the dashboard frontend whenever it changes.
type Status struct {
	Version            string
	UpdateAvailable    string
	CaptureRunning     bool
	CaptureError       bool
	ServerID           int
	IngestBaseURL      string
	CustomPublicIngest bool
	DriverWarning      string
	DriverHelpURL      string
	// EncryptionStatus is "" (unknown - nothing observed yet this
	// session), "encrypted" (a market data response came back
	// encrypted), or "clear" (a market data response came back and
	// decoded normally). See client/albion_state.go's
	// ShouldNotifyMarketDataEncrypted for how "encrypted" is decided.
	EncryptionStatus string
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

// SetVersionInfo records the running client version and, if one is known,
// the version available to update to (empty string if none).
func SetVersionInfo(version, updateAvailable string) {
	statusMu.Lock()
	next := status
	next.Version = version
	next.UpdateAvailable = updateAvailable
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

// SetCaptureRunning records whether packet capture is currently active.
func SetCaptureRunning(running bool) {
	statusMu.Lock()
	next := status
	next.CaptureRunning = running
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

// SetCaptureError records that packet capture has stopped due to a fatal
// error and will not resume without a restart. Unlike CaptureRunning,
// this is not expected to be cleared during the process's lifetime once
// set - a fatal capture error is terminal for that run.
func SetCaptureError(failed bool) {
	statusMu.Lock()
	next := status
	next.CaptureError = failed
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

// SetServer records the Albion Online server detected from captured
// traffic and the ingest URL that maps to it.
func SetServer(serverID int, ingestBaseURL string) {
	statusMu.Lock()
	next := status
	next.ServerID = serverID
	next.IngestBaseURL = ingestBaseURL
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

// SetCustomPublicIngest records whether the user has configured a custom
// (non-default) public ingest URL via -i, meaning uploads go to a
// self-hosted server instead of the Albion Data Project's public
// endpoint - not to be confused with ConfigGlobal.PrivateIngestBaseUrls
// (-p), a separate, unrelated feature for uploading a private copy of
// data to a second destination.
func SetCustomPublicIngest(custom bool) {
	statusMu.Lock()
	next := status
	next.CustomPublicIngest = custom
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

// SetDriverWarning records a problem with the system's packet-capture
// driver that the user should act on (see internal/pcapdriver), or
// clears it by passing empty strings.
func SetDriverWarning(message, helpURL string) {
	statusMu.Lock()
	next := status
	next.DriverWarning = message
	next.DriverHelpURL = helpURL
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

// EncryptionStatus values for SetEncryptionStatus.
const (
	EncryptionUnknown  = ""
	EncryptionDetected = "encrypted"
	EncryptionClear    = "clear"
)

// SetEncryptionStatus records whether market data has most recently come
// back encrypted, come back clear, or is unknown (EncryptionUnknown,
// the zero value - nothing observed yet, or reset because capture went
// idle and any prior observation is stale).
func SetEncryptionStatus(encryptionStatus string) {
	statusMu.Lock()
	next := status
	next.EncryptionStatus = encryptionStatus
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

// resetStatusForTest clears all package state. Test-only.
func resetStatusForTest() {
	statusMu.Lock()
	status = Status{}
	statusEmit = nil
	statusMu.Unlock()
}
