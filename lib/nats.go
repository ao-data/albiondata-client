package lib

const (
	// Public Topics
	NatsGoldPricesIngest       = "goldprices.ingest"
	NatsGoldPricesDeduped      = "goldprices.deduped"
	NatsMarketOrdersIngest     = "marketorders.ingest"
	NatsMarketOrdersDeduped    = "marketorders.deduped"
	NatsMarketHistoriesIngest  = "markethistories.ingest"
	NatsMarketHistoriesDeduped = "markethistories.deduped"
	NatsValidMarketOrders      = "validmarketorders"
	NatsMapDataIngest          = "mapdata.ingest"
	NatsMapDataDeduped         = "mapdata.deduped"
	NatsBanditEvent            = "banditevent.ingest"

	// Private Topics
	NatsSkillData           = "skills"
	NatsMarketNotifications = "marketnotifications"
)
