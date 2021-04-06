package types

import (
	"fmt"
	"time"
)

const (
	ModuleName = "poolyield"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	LockableDurationsKey = []byte("lockable_durations")
	DistrInfoKey         = []byte("distr_info")
)

func GetPoolFarmIdStoreKey(poolId uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("pool-yield/%d/%s", poolId, duration.String()))
}
