package types

import (
	"fmt"
)

const (
	ModuleName = "gamm"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	PoolAddressPrefix = []byte("gmm_liquidity_pool")
	GlobalPoolNumber  = []byte("gmm_global_pool_number")
)

func GetPoolShareDenom(poolId uint64) string {
	return fmt.Sprintf("gamm/pool/%d", poolId)
}
