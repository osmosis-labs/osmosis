package cfmm_common_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCalcExitPool(t *testing.T) {
	var (
		two_asset_pool = &stableswap.Pool{
			TotalShares:   sdk.NewCoin(gammtypes.GetPoolShareDenom(1), gammtypes.InitPoolSharesSupply),
			PoolLiquidity: sdk.NewCoins(sdk.NewInt64Coin("foo", 1000000000), sdk.NewInt64Coin("bar", 1000000000)),
		}

		three_asset_pool = &balancer.Pool{
			TotalShares: sdk.NewCoin(gammtypes.GetPoolShareDenom(1), sdk.NewIntFromUint64(6000000000)),
			PoolAssets: []balancer.PoolAsset{
				{Token: sdk.NewInt64Coin("foo", 2000000000)},
				{Token: sdk.NewInt64Coin("bar", 3000000000)},
				{Token: sdk.NewInt64Coin("baz", 4000000000)},
			},
		}
	)

	tests := []struct {
		name          string
		pool          gammtypes.PoolI
		exitingShares sdk.Int
		exitFee       sdk.Dec
		expExitCoins  sdk.Coins
		expError      bool
	}{
		{
			name:          "two-asset pool, zero exiting shares",
			pool:          two_asset_pool,
			exitingShares: sdk.NewIntFromUint64(0),
			exitFee:       sdk.ZeroDec(),
			expExitCoins:  sdk.Coins{},
			expError:      false,
		},
		{
			name:          "three-asset pool, zero exiting shares",
			pool:          three_asset_pool,
			exitingShares: sdk.NewIntFromUint64(0),
			exitFee:       sdk.ZeroDec(),
			expExitCoins:  sdk.Coins{},
			expError:      false,
		},
		{
			name:          "two-asset pool, exiting shares grater than total shares",
			pool:          two_asset_pool,
			exitingShares: gammtypes.InitPoolSharesSupply.AddRaw(1),
			exitFee:       sdk.ZeroDec(),
			expExitCoins:  sdk.Coins{},
			expError:      true,
		},
		{
			name:          "three-asset pool, exiting shares grater than total shares",
			pool:          three_asset_pool,
			exitingShares: sdk.NewIntFromUint64(7000000000),
			exitFee:       sdk.ZeroDec(),
			expExitCoins:  sdk.Coins{},
			expError:      true,
		},
		{
			name:          "two-asset pool, valid exiting shares",
			pool:          two_asset_pool,
			exitingShares: gammtypes.InitPoolSharesSupply.QuoRaw(2),
			exitFee:       sdk.ZeroDec(),
			expExitCoins:  sdk.Coins{sdk.NewInt64Coin("foo", 500000000), sdk.NewInt64Coin("bar", 500000000)},
			expError:      false,
		},
		{
			name:          "three-asset pool, valid exiting shares",
			pool:          three_asset_pool,
			exitingShares: sdk.NewIntFromUint64(3000000000),
			exitFee:       sdk.ZeroDec(),
			expExitCoins:  sdk.Coins{sdk.NewInt64Coin("foo", 1000000000), sdk.NewInt64Coin("bar", 1500000000), sdk.NewInt64Coin("baz", 2000000000)},
			expError:      false,
		},
		{
			name:          "two-asset pool with exit fee, valid exiting shares",
			pool:          two_asset_pool,
			exitingShares: gammtypes.InitPoolSharesSupply.QuoRaw(2),
			exitFee:       sdk.MustNewDecFromStr("0.0002"),
			expExitCoins:  sdk.Coins{sdk.NewInt64Coin("foo", 499900000), sdk.NewInt64Coin("bar", 499900000)},
			expError:      false,
		},
		{
			name:          "three-asset pool with exit fee, valid exiting shares",
			pool:          three_asset_pool,
			exitingShares: sdk.NewIntFromUint64(3000000000),
			exitFee:       sdk.MustNewDecFromStr("0.0001"),
			expExitCoins:  sdk.Coins{sdk.NewInt64Coin("foo", 999900000), sdk.NewInt64Coin("bar", 1499850000), sdk.NewInt64Coin("baz", 1999800000)},
			expError:      false,
		},
	}

	for _, test := range tests {
		// using empty context since it won't be used
		exitCoins, err := cfmm_common.CalcExitPool(sdk.Context{}, test.pool, test.exitingShares, test.exitFee)
		if test.expError {
			require.Error(t, err, "test: %v", test.name)
		} else {
			require.NoError(t, err, "test: %v", test.name)
			require.Equal(t, test.expExitCoins.Sort().String(), exitCoins.Sort().String(), "test: %v", test.name)
		}
	}
}
