package parser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/common"
	cosmwasmpool "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/model"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

func ConvertCFMM(ctx sdk.Context, pool poolmanagertypes.PoolI, bankKeeper common.BankKeeper) (domain.PoolI, error) {
	poolType := pool.GetType()
	if poolType != poolmanagertypes.Balancer && poolType != poolmanagertypes.Stableswap {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	balances := bankKeeper.GetAllBalances(ctx, pool.GetAddress())

	return &domain.Pool{
		UnderlyingPool: pool,
		// TODO: get it properly from TWAP and protorev routes
		TotalValueLockedUSDC: osmomath.OneInt(),
		Balances:             balances,
	}, nil
}

func ConvertConcentrated(ctx sdk.Context, pool poolmanagertypes.PoolI, bankeKeeper common.BankKeeper) (domain.PoolI, error) {
	if pool.GetType() != poolmanagertypes.Concentrated {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	balances := bankeKeeper.GetAllBalances(ctx, pool.GetAddress())

	return &domain.Pool{
		UnderlyingPool: pool,
		// TODO: get it properly from TWAP and protorev routes
		TotalValueLockedUSDC: osmomath.OneInt(),
		Balances:             balances,
	}, nil
}

func ConvertCosmWasm(ctx sdk.Context, pool poolmanagertypes.PoolI) (domain.PoolI, error) {
	if pool.GetType() != poolmanagertypes.CosmWasm {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	// Cast to concentrated pool
	cosmwasmPool, ok := pool.(*cosmwasmpool.Pool)
	if !ok {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	balances := cosmwasmPool.GetTotalPoolLiquidity(ctx)

	denoms := make([]string, 0, len(balances))
	for _, balance := range balances {
		denoms = append(denoms, balance.Denom)
	}

	return &domain.Pool{
		UnderlyingPool: pool,
		// TODO: get it properly from TWAP and protorev routes
		TotalValueLockedUSDC: osmomath.OneInt(),
		Balances:             balances,
	}, nil
}
