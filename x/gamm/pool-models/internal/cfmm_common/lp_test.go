package cfmm_common_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// a helper function used to multiply coins
func mulCoins(coins sdk.Coins, multiplier osmomath.Dec) sdk.Coins {
	outCoins := sdk.Coins{}
	for _, coin := range coins {
		outCoin := sdk.NewCoin(coin.Denom, multiplier.MulInt(coin.Amount).TruncateInt())
		if !outCoin.Amount.IsZero() {
			outCoins = append(outCoins, outCoin)
		}
	}
	return outCoins
}

func TestCalcExitPool(t *testing.T) {
	emptyContext := sdk.Context{}

	twoStablePoolAssets := sdk.NewCoins(
		sdk.NewInt64Coin("foo", 1000000000),
		sdk.NewInt64Coin("bar", 1000000000),
	)

	threeBalancerPoolAssets := []balancer.PoolAsset{
		{Token: sdk.NewInt64Coin("foo", 2000000000), Weight: osmomath.NewIntFromUint64(5)},
		{Token: sdk.NewInt64Coin("bar", 3000000000), Weight: osmomath.NewIntFromUint64(5)},
		{Token: sdk.NewInt64Coin("baz", 4000000000), Weight: osmomath.NewIntFromUint64(5)},
	}

	// create these pools used for testing
	twoAssetPool, err := stableswap.NewStableswapPool(
		1,
		stableswap.PoolParams{ExitFee: osmomath.ZeroDec()},
		twoStablePoolAssets,
		[]uint64{1, 1},
		"",
		"",
	)
	require.NoError(t, err)

	threeAssetPool, err := balancer.NewBalancerPool(
		1,
		balancer.PoolParams{SwapFee: osmomath.ZeroDec(), ExitFee: osmomath.ZeroDec()},
		threeBalancerPoolAssets,
		"",
		time.Now(),
	)
	require.NoError(t, err)

	twoAssetPoolWithExitFee, err := stableswap.NewStableswapPool(
		1,
		stableswap.PoolParams{ExitFee: osmomath.MustNewDecFromStr("0.0001")},
		twoStablePoolAssets,
		[]uint64{1, 1},
		"",
		"",
	)
	require.NoError(t, err)

	threeAssetPoolWithExitFee, err := balancer.NewBalancerPool(
		1,
		balancer.PoolParams{SwapFee: osmomath.ZeroDec(), ExitFee: osmomath.MustNewDecFromStr("0.0002")},
		threeBalancerPoolAssets,
		"",
		time.Now(),
	)
	require.NoError(t, err)

	tests := []struct {
		name          string
		pool          gammtypes.CFMMPoolI
		exitingShares osmomath.Int
		expError      bool
	}{
		{
			name:          "two-asset pool, exiting shares grater than total shares",
			pool:          &twoAssetPool,
			exitingShares: twoAssetPool.GetTotalShares().AddRaw(1),
			expError:      true,
		},
		{
			name:          "three-asset pool, exiting shares grater than total shares",
			pool:          &threeAssetPool,
			exitingShares: threeAssetPool.GetTotalShares().AddRaw(1),
			expError:      true,
		},
		{
			name:          "two-asset pool, valid exiting shares",
			pool:          &twoAssetPool,
			exitingShares: twoAssetPool.GetTotalShares().QuoRaw(2),
			expError:      false,
		},
		{
			name:          "three-asset pool, valid exiting shares",
			pool:          &threeAssetPool,
			exitingShares: osmomath.NewIntFromUint64(3000000000000),
			expError:      false,
		},
		{
			name:          "two-asset pool with exit fee, valid exiting shares",
			pool:          &twoAssetPoolWithExitFee,
			exitingShares: twoAssetPoolWithExitFee.GetTotalShares().QuoRaw(2),
			expError:      false,
		},
		{
			name:          "three-asset pool with exit fee, valid exiting shares",
			pool:          &threeAssetPoolWithExitFee,
			exitingShares: osmomath.NewIntFromUint64(7000000000000),
			expError:      false,
		},
	}

	for _, test := range tests {
		// using empty context since, currently, the context is not used anyway. This might be changed in the future
		exitFee := test.pool.GetExitFee(emptyContext)
		exitCoins, err := cfmm_common.CalcExitPool(emptyContext, test.pool, test.exitingShares, exitFee)
		if test.expError {
			require.Error(t, err, "test: %v", test.name)
		} else {
			require.NoError(t, err, "test: %v", test.name)

			// exitCoins = ( (1 - exitFee) * exitingShares / poolTotalShares ) * poolTotalLiquidity
			expExitCoins := mulCoins(test.pool.GetTotalPoolLiquidity(emptyContext), (osmomath.OneDec().Sub(exitFee)).MulInt(test.exitingShares).QuoInt(test.pool.GetTotalShares()))
			require.Equal(t, expExitCoins.Sort().String(), exitCoins.Sort().String(), "test: %v", test.name)
		}
	}
}

func TestMaximalExactRatioJoin(t *testing.T) {
	emptyContext := sdk.Context{}

	balancerPoolAsset := []balancer.PoolAsset{
		{Token: sdk.NewInt64Coin("foo", 100), Weight: osmomath.NewIntFromUint64(5)},
		{Token: sdk.NewInt64Coin("bar", 100), Weight: osmomath.NewIntFromUint64(5)},
	}

	tests := []struct {
		name        string
		pool        func() poolmanagertypes.PoolI
		tokensIn    sdk.Coins
		expNumShare osmomath.Int
		expRemCoin  sdk.Coins
	}{
		{
			name: "two asset pool, same tokenIn ratio",
			pool: func() poolmanagertypes.PoolI {
				balancerPool, err := balancer.NewBalancerPool(
					1,
					balancer.PoolParams{SwapFee: osmomath.ZeroDec(), ExitFee: osmomath.ZeroDec()},
					balancerPoolAsset,
					"",
					time.Now(),
				)
				require.NoError(t, err)
				return &balancerPool
			},
			tokensIn:    sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10)), sdk.NewCoin("bar", osmomath.NewInt(10))),
			expNumShare: osmomath.NewIntFromUint64(10000000000000000000),
			expRemCoin:  sdk.Coins{},
		},
		{
			name: "two asset pool, different tokenIn ratio with pool",
			pool: func() poolmanagertypes.PoolI {
				balancerPool, err := balancer.NewBalancerPool(
					1,
					balancer.PoolParams{SwapFee: osmomath.ZeroDec(), ExitFee: osmomath.ZeroDec()},
					balancerPoolAsset,
					"",
					time.Now(),
				)
				require.NoError(t, err)
				return &balancerPool
			},
			tokensIn:    sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(10)), sdk.NewCoin("bar", osmomath.NewInt(11))),
			expNumShare: osmomath.NewIntFromUint64(10000000000000000000),
			expRemCoin:  sdk.NewCoins(sdk.NewCoin("bar", osmomath.NewIntFromUint64(1))),
		},
	}

	for _, test := range tests {
		balancerPool, err := balancer.NewBalancerPool(
			1,
			balancer.PoolParams{SwapFee: osmomath.ZeroDec(), ExitFee: osmomath.ZeroDec()},
			balancerPoolAsset,
			"",
			time.Now(),
		)
		require.NoError(t, err)

		numShare, remCoins, err := cfmm_common.MaximalExactRatioJoin(&balancerPool, emptyContext, test.tokensIn)

		require.NoError(t, err)
		require.Equal(t, test.expNumShare, numShare)
		require.Equal(t, test.expRemCoin, remCoins)
	}
}
