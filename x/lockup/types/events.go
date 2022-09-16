package types

// event types.
const (
	TypeEvtLockTokens      = "lockup_lock_tokens"
	TypeEvtAddTokensToLock = "lockup_add_tokens_to_lock"
	TypeEvtBeginUnlockAll  = "lockup_begin_unlock_all"
	TypeEvtBeginUnlock     = "lockup_begin_unlock"

	AttributePeriodLockID         = "period_lock_id"
	AttributePeriodLockOwner      = "owner"
	AttributePeriodLockAmount     = "amount"
	AttributePeriodLockDuration   = "duration"
	AttributePeriodLockUnlockTime = "unlock_time"
	AttributeUnlockedCoins        = "unlocked_coins"
)
