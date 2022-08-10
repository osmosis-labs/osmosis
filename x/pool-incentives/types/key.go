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

// GetPoolGaugeIdStoreKey returns a StoreKey with pool ID and its duration as inputs
func GetPoolGaugeIdStoreKey(poolId uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-incentives/%d/%s", poolId, duration.String()))
}

// GetPoolIdFromGaugeIdStoreKey returns a StoreKey from the given gaugeID and duration.
func GetPoolIdFromGaugeIdStoreKey(gaugeId uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-incentives-pool-id/%d/%s", gaugeId, duration.String()))
}
