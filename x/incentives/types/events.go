package types

// Incentive module event types.
const (
	TypeEvtCreateGauge  = "incentives_create_gauge"
	TypeEvtAddToGauge   = "incentives_add_to_gauge"
	TypeEvtDistribution = "incentives_distribution"

	AttributeGaugeID     = "gauge_id"
	AttributeLockedDenom = "denom"
	AttributeReceiver    = "receiver"
	AttributeAmount      = "amount"
)
