package pools

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	concentratedmodel "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/model"
	cwpoolmodel "github.com/osmosis-labs/osmosis/v21/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v21/x/gamm/pool-models/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// NewRoutablePool creates a new RoutablePool.
// Panics if pool is of invalid type or if does not contain tick data when a concentrated pool.
func NewRoutablePool(pool domain.PoolI, tokenOutDenom string, takerFee osmomath.Dec) domain.RoutablePool {
	poolType := pool.GetType()
	chainPool := pool.GetUnderlyingPool()
	if poolType == poolmanagertypes.Concentrated {
		// Check if pools is concentrated
		concentratedPool, ok := chainPool.(*concentratedmodel.Pool)
		if !ok {
			panic(domain.FailedToCastPoolModelError{
				ExpectedModel: poolmanagertypes.PoolType_name[int32(poolmanagertypes.Concentrated)],
				ActualModel:   poolmanagertypes.PoolType_name[int32(poolType)],
			})
		}

		tickModel, err := pool.GetTickModel()
		if err != nil {
			panic(err)
		}

		return &routableConcentratedPoolImpl{
			ChainPool:     concentratedPool,
			TickModel:     tickModel,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee,
		}
	}

	if poolType == poolmanagertypes.Balancer {
		chainPool := pool.GetUnderlyingPool()

		// Check if pools is balancer
		balancerPool, ok := chainPool.(*balancer.Pool)
		if !ok {
			panic(domain.FailedToCastPoolModelError{
				ExpectedModel: poolmanagertypes.PoolType_name[int32(poolmanagertypes.Balancer)],
				ActualModel:   poolmanagertypes.PoolType_name[int32(poolType)],
			})
		}

		return &routableBalancerPoolImpl{
			ChainPool:     balancerPool,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee,
		}
	}

	if pool.GetType() == poolmanagertypes.CosmWasm {
		cosmwasmPool, ok := chainPool.(*cwpoolmodel.CosmWasmPool)
		if !ok {
			panic(domain.FailedToCastPoolModelError{
				ExpectedModel: poolmanagertypes.PoolType_name[int32(poolmanagertypes.Balancer)],
				ActualModel:   poolmanagertypes.PoolType_name[int32(poolType)],
			})
		}

		sqsPoolModel := pool.GetSQSPoolModel().SpreadFactor

		return &routableTransmuterPoolImpl{
			ChainPool:     cosmwasmPool,
			Balances:      pool.GetSQSPoolModel().Balances,
			TokenOutDenom: tokenOutDenom,
			TakerFee:      takerFee,
			SpreadFactor:  sqsPoolModel,
		}
	}

	// Must be stableswap
	if poolType != poolmanagertypes.Stableswap {
		panic(domain.InvalidPoolTypeError{
			PoolType: int32(poolType),
		})
	}

	// Check if pools is balancer
	stableswapPool, ok := chainPool.(*stableswap.Pool)
	if !ok {
		panic(domain.FailedToCastPoolModelError{
			ExpectedModel: poolmanagertypes.PoolType_name[int32(poolmanagertypes.Balancer)],
			ActualModel:   poolmanagertypes.PoolType_name[int32(poolType)],
		})
	}

	return &routableStableswapPoolImpl{
		ChainPool:     stableswapPool,
		TokenOutDenom: tokenOutDenom,
		TakerFee:      takerFee,
	}
}
