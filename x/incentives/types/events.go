package types

// Incentive module event types.
const (
	TypeEvtCreateGauge  = "create_gauge"
	TypeEvtAddToGauge   = "add_to_gauge"
	TypeEvtDistribution = "distribution"

	AttributeGaugeID     = "gauge_id"
	AttributeLockedDenom = "denom"
	AttributeReceiver    = "receiver"
	AttributeAmount      = "amount"
)
