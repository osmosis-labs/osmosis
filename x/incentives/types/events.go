package types

// event types
const (
	TypeEvtCreateGauge  = "create_gauge"
	TypeEvtAddToGauge   = "add_to_gauge"
	TypeEvtDistribution = "distribution"

	AttributeGaugeID  = "gauge_id"
	AttributeLockID   = "lock_id"
	AttributeReceiver = "receiver"
	AttributeAmount   = "amount"
)
