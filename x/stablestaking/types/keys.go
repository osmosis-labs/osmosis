package types

const (
	// ModuleName defines the module name
	ModuleName = "stablestaking"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for stablestaking
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_stablestaking"

	// ParamsKey is the key for params
	ParamsKey = "params"

	// PoolKey is the key for pools
	PoolKey = "pool"

	// UserStakeKey is the key for user stakes
	UserStakeKey = "user_stake"

	// SnapshotKey is the key for snapshots
	SnapshotKey = "snapshot"

	// UnbondingKey is the key for unbonding info
	UnbondingKey = "unbonding"

	// NativeRewardsCollectorName is the key for native rewards collector
	NativeRewardsCollectorName = "native_rewards_collector"
)
