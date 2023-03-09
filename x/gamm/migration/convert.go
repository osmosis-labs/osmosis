package migration

import (
	oldbalancer "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	oldstableswap "github.com/osmosis-labs/osmosis/v15/x/gamm/v2types/stableswap"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"
)

func convertToNewBalancerPool(oldPool oldbalancer.Pool) balancer.Pool {
	return balancer.Pool{
		Address: oldPool.Address,
		Id: oldPool.Id,
		PoolParams: balancer.PoolParams{
			SwapFee: oldPool.PoolParams.SwapFee,
			SmoothWeightChangeParams: oldPool.PoolParams.SmoothWeightChangeParams,
		},
		FuturePoolGovernor: oldPool.FuturePoolGovernor,
		TotalShares: oldPool.TotalShares,
		PoolAssets: oldPool.PoolAssets,
		TotalWeight: oldPool.TotalWeight,
	}
}

func convertToNewStableSwapPool(oldPool oldstableswap.Pool) stableswap.Pool {
	return stableswap.Pool{
		Address: oldPool.Address,
		Id: oldPool.Id,
		PoolParams: stableswap.PoolParams{
			SwapFee: oldPool.PoolParams.SwapFee,
		},
		FuturePoolGovernor: oldPool.FuturePoolGovernor,
		TotalShares: oldPool.TotalShares,
		PoolLiquidity: oldPool.PoolLiquidity,
		ScalingFactors: oldPool.ScalingFactors,
		ScalingFactorController: oldPool.ScalingFactorController,
	}
}