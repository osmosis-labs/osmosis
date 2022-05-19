package cfmm_common_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// a helper function used to multiply coins
func mulCoins(coins sdk.Coins, multiplier sdk.Dec) sdk.Coins {
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
	var (
		emptyContext = sdk.Context{}

		// create these pools used for testing
		twoAssetPool, err1 = stableswap.NewStableswapPool(
			1,
			stableswap.PoolParams{ExitFee: sdk.ZeroDec()},
			sdk.NewCoins(sdk.NewInt64Coin("foo", 1000000000),
				sdk.NewInt64Coin("bar", 1000000000)),
			"",
			time.Now(),
		)
		threeAssetPool, err2 = balancer.NewBalancerPool(
			1,
			balancer.PoolParams{SwapFee: sdk.ZeroDec(), ExitFee: sdk.ZeroDec()},
			[]balancer.PoolAsset{
				{Token: sdk.NewInt64Coin("foo", 2000000000), Weight: sdk.NewIntFromUint64(5)},
				{Token: sdk.NewInt64Coin("bar", 3000000000), Weight: sdk.NewIntFromUint64(5)},
				{Token: sdk.NewInt64Coin("baz", 4000000000), Weight: sdk.NewIntFromUint64(5)},
			},
			"",
			time.Now(),
		)
		twoAssetPoolWithExitFee, err3 = stableswap.NewStableswapPool(
			1,
			stableswap.PoolParams{ExitFee: sdk.MustNewDecFromStr("0.0001")},
			sdk.NewCoins(sdk.NewInt64Coin("foo", 1000000000),
				sdk.NewInt64Coin("bar", 1000000000)),
			"",
			time.Now(),
		)
		threeAssetPoolWithExitFee, err4 = balancer.NewBalancerPool(
			1,
			balancer.PoolParams{SwapFee: sdk.ZeroDec(), ExitFee: sdk.MustNewDecFromStr("0.0002")},
			[]balancer.PoolAsset{
				{Token: sdk.NewInt64Coin("foo", 2000000000), Weight: sdk.NewIntFromUint64(5)},
				{Token: sdk.NewInt64Coin("bar", 3000000000), Weight: sdk.NewIntFromUint64(5)},
				{Token: sdk.NewInt64Coin("baz", 4000000000), Weight: sdk.NewIntFromUint64(5)},
			},
			"",
			time.Now(),
		)
	)

	// make sure there're no error creating those pools
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	require.NoError(t, err4)

	tests := []struct {
		name          string
		pool          gammtypes.PoolI
		exitingShares sdk.Int
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
			exitingShares: sdk.NewIntFromUint64(3000000000000),
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
			exitingShares: sdk.NewIntFromUint64(7000000000000),
			expError:      false,
		},
	}

	for _, test := range tests {
		// using empty context since the context won't be used anyway
		exitFee := test.pool.GetExitFee(emptyContext)
		exitCoins, err := cfmm_common.CalcExitPool(emptyContext, test.pool, test.exitingShares, exitFee)
		if test.expError {
			require.Error(t, err, "test: %v", test.name)
		} else {
			require.NoError(t, err, "test: %v", test.name)

			// exitCoins = ( (1 - exitFee) * exitingShares / poolTotalShares ) * poolTotalLiquidity
			expExitCoins := mulCoins(test.pool.GetTotalPoolLiquidity(emptyContext), (sdk.OneDec().Sub(exitFee)).MulInt(test.exitingShares).QuoInt(test.pool.GetTotalShares()))
			require.Equal(t, expExitCoins.Sort().String(), exitCoins.Sort().String(), "test: %v", test.name)
		}
	}
}
