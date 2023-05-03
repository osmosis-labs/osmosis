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

	TypeEvtUnlockAndMigrateShares               = "unlock_and_migrate_shares"
	TypeEvtCreateFullRangePositionAndSFDelegate = "full_range_position_and_delegate"
	AttributeKeyPoolIdEntering                  = "pool_id_entering"
	AttributeKeyPoolIdLeaving                   = "pool_id_leaving"
	AttributeGammLockId                         = "gamm_lock_id"
	AttributeConcentratedLockId                 = "concentrated_lock_id"
	AttributeKeyPoolId                          = "pool_id"
	AttributePositionId                         = "position_id"
	AttributeAmount0                            = "amount0"
	AttributeAmount1                            = "amount1"
	AttributeLiquidity                          = "liquidity"
	AttributeJoinTime                           = "join_time"

	AttributeDenom               = "denom"
	AttributeSuperfluidAssetType = "superfluid_asset_type"
	AttributeLockId              = "lock_id"
	AttributeValidator           = "validator"
	AttributeAmount              = "amount"
)
