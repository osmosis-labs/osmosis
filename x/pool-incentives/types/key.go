package types

import (
	"fmt"
	"time"
)

const (
	ModuleName = "poolincentives"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	LockableDurationsKey = []byte("lockable_durations")
	DistrInfoKey         = []byte("distr_info")
)

// GetPoolGaugeIdInternalStoreKey returns a StoreKey with pool ID and its duration as inputs
// This is used for linking pool id, duration and gauge id for internal incentives.
func GetPoolGaugeIdInternalStoreKey(poolId uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-incentives/%d/%s", poolId, duration.String()))
}

// GetPoolIdFromGaugeIdStoreKey returns a StoreKey from the given gaugeID and duration.
func GetPoolIdFromGaugeIdStoreKey(gaugeId uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-incentives-pool-id/%d/%s", gaugeId, duration.String()))
}

// GetPoolNoLockGaugeIdStoreKey returns a StoreKey with pool ID and gauge id as input
// assumming that the pool has no lockable duration.
func GetPoolNoLockGaugeIdStoreKey(poolId uint64, gaugeId uint64) []byte {
	return []byte(fmt.Sprintf("no-lock-pool-incentives/%d/%d", poolId, gaugeId))
}

// GetPoolNoLockGaugeIdIterationStoreKey returns a StoreKey with pool ID as input
// assumming that the pool has no lockable duration. It is used for collecting
// values by iterating over this prefix.
func GetPoolNoLockGaugeIdIterationStoreKey(poolId uint64) []byte {
	return []byte(fmt.Sprintf("no-lock-pool-incentives/%d/", poolId))
}
