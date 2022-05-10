package stableswap_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
	types "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

const (
	usdaDenom = "usda"
	usdbDenom = "usdb"
	usdcDenom = "usdc"
	usdtDenom = "usdt"
	ustDenom  = "ust"
)

func createTestPool(t *testing.T, poolId uint64, poolAssets []stableswap.PoolAsset, swapFee, exitFee sdk.Dec) types.PoolI {
	pool, err := stableswap.NewStableswapPool(poolId, stableswap.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	},
		poolAssets,
		"")

	require.NoError(t, err)
	require.NotNil(t, pool)

	return pool
}

func createTestPoolNoValidation(t *testing.T, poolId uint64, poolAssets []stableswap.PoolAsset, swapFee, exitFee sdk.Dec) types.PoolI {
	pool := stableswap.Pool{
		Address: types.NewPoolAddress(poolId).String(),
		Id:      poolId,
		PoolParams: stableswap.PoolParams{
			SwapFee: swapFee,
			ExitFee: exitFee,
		},
		TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(poolId), types.InitPoolSharesSupply),
		PoolAssets:         poolAssets,
		FuturePoolGovernor: "",
	}
	return &pool
}
