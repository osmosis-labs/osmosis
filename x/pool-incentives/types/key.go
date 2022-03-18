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

func GetPoolGaugeIDStoreKey(poolID uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-incentives/%d/%s", poolID, duration.String()))
}

func GetPoolIDFromGaugeIDStoreKey(gaugeID uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-incentives-pool-id/%d/%s", gaugeID, duration.String()))
}
