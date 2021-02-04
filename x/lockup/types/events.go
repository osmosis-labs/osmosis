package types

// event types
const (
	TypeEvtLockTokens   = "lock_tokens"
	TypeEvtUnlockTokens = "unlock_tokens"
	TypeEvtUnlock       = "unlock"

	AttributePeriodLockID         = "period_lock_id"
	AttributePeriodLockOwner      = "owner"
	AttributePeriodLockAmount     = "amount"
	AttributePeriodLockDuration   = "duration"
	AttributePeriodLockUnlockTime = "unlock_time"
	AttributeUnlockedCoins        = "unlocked_coins"
)
