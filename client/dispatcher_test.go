package client

import (
	"testing"

	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/lib"
)

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
