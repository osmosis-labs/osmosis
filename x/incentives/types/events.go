package types

// Incentive module event types.
const (
	TypeEvtCreateGauge  = "create_gauge"
	TypeEvtAddToGauge   = "add_to_gauge"
	TypeEvtCreateGroup  = "create_group"
	TypeEvtDistribution = "distribution"

	AttributeGaugeID     = "gauge_id"
	AttributeGroupID     = "group_id"
	AttributeLockedDenom = "denom"
	AttributeReceiver    = "receiver"
	AttributeAmount      = "amount"
)
