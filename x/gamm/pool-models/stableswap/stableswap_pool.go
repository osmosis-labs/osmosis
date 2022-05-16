package stableswap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v8/x/gamm/types"
)

var _ types.PoolI = &Pool{}

// NewStableswapPool returns a stableswap pool
// Invariants that are assumed to be satisfied and not checked:
// * len(initialLiquidity) = 2
// * FutureGovernor is valid
// * poolID doesn't already exist
func NewStableswapPool(poolId uint64, stableswapPoolParams PoolParams, initialLiquidity sdk.Coins, futureGovernor string, blockTime time.Time) (Pool, error) {
	pool := Pool{
		Address:            types.NewPoolAddress(poolId).String(),
		Id:                 poolId,
		PoolParams:         stableswapPoolParams,
		TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply),
		PoolLiquidity:      initialLiquidity,
		FuturePoolGovernor: futureGovernor,
	}

	return pool, nil
}
