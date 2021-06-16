package types

// event types
const (
	TypeEvtLockTokens     = "lock_tokens"
	TypeEvtBeginUnlockAll = "begin_unlock_all"
	TypeEvtUnlockTokens   = "unlock_tokens"
	TypeEvtBeginUnlock    = "begin_unlock"
	TypeEvtUnlock         = "unlock"

	AttributePeriodLockID         = "period_lock_id"
	AttributePeriodLockOwner      = "owner"
	AttributePeriodLockAmount     = "amount"
	AttributePeriodLockDuration   = "duration"
	AttributePeriodLockUnlockTime = "unlock_time"
	AttributeUnlockedCoins        = "unlocked_coins"
)
