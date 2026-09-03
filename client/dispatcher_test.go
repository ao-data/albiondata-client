package client

import (
	"testing"

	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/lib"
)

// TestCreateUploaders_ReusesUploaderForSameTarget guards against the
// memory leak where every upload call constructed brand new uploaders
// (a fresh NATS connection that was never closed, a fresh http.Transport
// whose idle connections were never reclaimed) instead of reusing one
// per ingest target.
func TestCreateUploaders_ReusesUploaderForSameTarget(t *testing.T) {
	target := "https://example.com/TestCreateUploaders_ReusesUploaderForSameTarget"

	first := createUploaders([]string{target})
	second := createUploaders([]string{target})

	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("got %d and %d uploaders, want 1 and 1", len(first), len(second))
	}
	if first[0] != second[0] {
		t.Fatal("createUploaders returned a different uploader instance for the same target")
	}
}

func TestSendMsgToPublicUploaders_IncrementsCounterByRecordCount(t *testing.T) {
	// No configured ingest targets: createUploaders returns nothing for
	// both public and private, so this exercises the counter increment
	// without making any network calls.
	ConfigGlobal.PublicIngestBaseUrls = ""
	ConfigGlobal.PrivateIngestBaseUrls = ""

	before := dashboard.GetUploadCounts()["marketorders.ingest"]

	sendMsgToPublicUploaders(struct{}{}, lib.NatsMarketOrdersIngest, &albionState{}, "test-id", 50)

	after := dashboard.GetUploadCounts()["marketorders.ingest"]
	if after != before+50 {
		t.Fatalf("marketorders.ingest counter = %d, want %d", after, before+50)
	}
}
