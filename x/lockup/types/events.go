package types

// event types
const (
	TypeEvtLockTokens     = "lock_tokens"
	TypeEvtBeginUnlockAll = "begin_unlock_all"
	TypeEvtBeginUnlock    = "begin_unlock"

	AttributePeriodLockID         = "period_lock_id"
	AttributePeriodLockOwner      = "owner"
	AttributePeriodLockAmount     = "amount"
	AttributePeriodLockDuration   = "duration"
	AttributePeriodLockUnlockTime = "unlock_time"
	AttributeUnlockedCoins        = "unlocked_coins"
)
