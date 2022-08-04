package types

// event types.
const (
	TypeEvtSetSuperfluidAsset           = "set_superfluid_asset"
	TypeEvtRemoveSuperfluidAsset        = "remove_superfluid_asset"
	TypeEvtSuperfluidDelegate           = "superfluid_delegate"
	TypeEvtSuperfluidIncreaseDelegation = "superfluid_increase_delegation"
	TypeEvtSuperfluidUndelegate         = "superfluid_undelegate"
	TypeEvtSuperfluidUnbondLock         = "superfluid_unbond_lock"

	TypeEvtUnpoolId     = "unpool_pool_id"
	AttributeNewLockIds = "new_lock_ids"

	AttributeDenom               = "denom"
	AttributeSuperfluidAssetType = "superfluid_asset_type"
	AttributeLockId              = "lock_id"
	AttributeValidator           = "validator"
	AttributeAmount              = "amount"
)
