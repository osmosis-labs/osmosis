package types

// event types.
const (
	TypeEvtSetSuperfluidAsset           = "superfluid_set_asset"
	TypeEvtRemoveSuperfluidAsset        = "superfluid_remove_asset"
	TypeEvtSuperfluidDelegate           = "superfluid_delegate"
	TypeEvtSuperfluidIncreaseDelegation = "superfluid_increase_delegation"
	TypeEvtSuperfluidUndelegate         = "superfluid_undelegate"
	TypeEvtSuperfluidUnbondLock         = "superfluid_unbond_lock"

	TypeEvtUnpoolId     = "superfluid_unpool_pool_id"
	AttributeNewLockIds = "new_lock_ids"

	AttributeDenom               = "denom"
	AttributeSuperfluidAssetType = "superfluid_asset_type"
	AttributeLockId              = "lock_id"
	AttributeValidator           = "validator"
	AttributeAmount              = "amount"
)
