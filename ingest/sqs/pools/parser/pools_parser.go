package parser

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/common"
	concentrated "github.com/osmosis-labs/osmosis/v20/x/concentrated-liquidity/model"
	cosmwasmpool "github.com/osmosis-labs/osmosis/v20/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v20/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v20/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v20/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

func ConvertCFMM(ctx sdk.Context, pool poolmanagertypes.PoolI) (domain.PoolI, error) {
	poolType := pool.GetType()
	if poolType != poolmanagertypes.Balancer && poolType != poolmanagertypes.Stableswap {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	// Cast to CFMM pool
	cfmmPool, ok := pool.(gammtypes.CFMMPoolI)
	if !ok {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	totalPoolBalances := cfmmPool.GetTotalPoolLiquidity(ctx)
	poolDenoms := make([]string, 0, len(totalPoolBalances))
	poolWeights := make([]string, 0)
	for i, balance := range totalPoolBalances {
		poolDenoms = append(poolDenoms, balance.Denom)

		// TODO: figure out if there is a cleaner way to do this.
		if poolType == poolmanagertypes.Balancer {
			balancerPool, ok := cfmmPool.(*balancer.Pool)
			if !ok {
				return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
			}
			poolWeights = append(poolWeights, balancerPool.PoolAssets[i].Weight.String())
		} else {
			stableswapPool, ok := cfmmPool.(*stableswap.Pool)
			if !ok {
				return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
			}
			poolWeights = append(poolWeights, strconv.FormatUint(stableswapPool.ScalingFactors[i], 10))
		}
	}

	return &domain.Pool{
		Id:           cfmmPool.GetId(),
		Type:         int(pool.GetType()),
		Liquidity:    cfmmPool.GetTotalShares().String(),
		SpreadFactor: cfmmPool.GetSpreadFactor(ctx).String(),
		Balances:     totalPoolBalances.String(),
		Denoms:       poolDenoms,
		Weights:      poolWeights,
	}, nil
}

func ConvertConcentrated(ctx sdk.Context, pool poolmanagertypes.PoolI, bankeKeeper common.BankKeeper) (domain.PoolI, error) {
	if pool.GetType() != poolmanagertypes.Concentrated {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	// Cast to concentrated pool
	concentratedPool, ok := pool.(*concentrated.Pool)
	if !ok {
		return nil, domain.InvalidPoolTypeError{PoolType: int32(pool.GetType())}
	}

	balances := bankeKeeper.GetAllBalances(ctx, concentratedPool.GetAddress())

	return &domain.Pool{
		Id:           concentratedPool.Id,
		Type:         int(pool.GetType()),
		Liquidity:    concentratedPool.CurrentTickLiquidity.String(),
		SpreadFactor: concentratedPool.SpreadFactor.String(),
		Balances:     balances.String(),
		Denoms:       []string{concentratedPool.Token0, concentratedPool.Token1},
		// No weights
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
		Id:           cosmwasmPool.PoolId,
		Type:         int(pool.GetType()),
		Balances:     balances.String(),
		SpreadFactor: cosmwasmPool.GetSpreadFactor(ctx).String(),
		Denoms:       denoms,
		// No liquidity
		// No weights
	}, nil
}
