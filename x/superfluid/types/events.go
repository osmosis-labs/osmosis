package types

// event types.
const (
	TypeEvtSetSuperfluidAsset                = "set_superfluid_asset"
	TypeEvtRemoveSuperfluidAsset             = "remove_superfluid_asset"
	TypeEvtSuperfluidDelegate                = "superfluid_delegate"
	TypeEvtSuperfluidIncreaseDelegation      = "superfluid_increase_delegation"
	TypeEvtSuperfluidUndelegate              = "superfluid_undelegate"
	TypeEvtSuperfluidUnbondLock              = "superfluid_unbond_lock"
	TypeEvtSuperfluidUndelegateAndUnbondLock = "superfluid_undelegate_and_unbond_lock"

	TypeEvtUnpoolId     = "unpool_pool_id"
	AttributeNewLockIds = "new_lock_ids"

	TypeEvtUnlockAndMigrateShares = "unlock_and_migrate_shares"
	AttributeKeyPoolIdEntering    = "pool_id_entering"
	AttributeKeyPoolIdLeaving     = "pool_id_leaving"
	AttributeNewLockId            = "new_lock_id"
	AttributeFreezeDuration       = "freeze_duration"
	AttributeJoinTime             = "join_time"

	AttributeDenom               = "denom"
	AttributeSuperfluidAssetType = "superfluid_asset_type"
	AttributeLockId              = "lock_id"
	AttributeValidator           = "validator"
	AttributeAmount              = "amount"
)
