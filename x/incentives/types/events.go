package types

// event types
const (
	TypeEvtCreateGauge        = "create_gauge"
	TypeEvtAddToGauge         = "add_to_gauge"
	TypeEvtDistribution       = "distribution"
	TypeEvtClaimLockReward    = "claim_lock_reward"
	TypeEvtClaimLockRewardAll = "claim_lock_reward_all"

	AttributeGaugeID     = "gauge_id"
	AttributeLockedDenom = "denom"
	AttributeReceiver    = "receiver"
	AttributeAmount      = "amount"

	AttributePeriodLockID    = "period_lock_id"
	AttributePeriodLockOwner = "owner"
	AttributeRewardCoins     = "reward_coins"
)
