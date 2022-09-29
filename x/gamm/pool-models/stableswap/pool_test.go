// nolint: composites
package stableswap

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

var (
	defaultSwapFee              = sdk.MustNewDecFromStr("0.025")
	defaultExitFee              = sdk.ZeroDec()
	defaultPoolId               = uint64(1)
	defaultStableswapPoolParams = PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultTwoAssetScalingFactors = []uint64{1, 1}
	defaultFutureGovernor         = ""

	twoEvenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("foo", 1000000000),
		sdk.NewInt64Coin("bar", 1000000000),
	)
	twoUnevenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("foo", 2000000000),
		sdk.NewInt64Coin("bar", 1000000000),
	)
	fiveUnevenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("asset/a", 1000000000),
		sdk.NewInt64Coin("asset/b", 2000000000),
		sdk.NewInt64Coin("asset/c", 3000000000),
		sdk.NewInt64Coin("asset/d", 4000000000),
		sdk.NewInt64Coin("asset/e", 5000000000),
	)
)

func TestGetScaledPoolAmts(t *testing.T) {
	tests := map[string]struct {
		denoms         []string
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expReserves    []sdk.Dec
		expPanic       bool
	}{
		// sanity checks, default scaling factors
		"get both pool assets, even two-asset pool with default scaling factors": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []sdk.Dec{sdk.NewDec(1000000000), sdk.NewDec(1000000000)},
			expPanic:       false,
		},
		"get one pool asset, even two-asset pool with default scaling factors": {
			denoms:         []string{"foo"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []sdk.Dec{sdk.NewDec(1000000000)},
			expPanic:       false,
		},
		"get both pool assets, uneven two-asset pool with default scaling factors": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []sdk.Dec{sdk.NewDec(2000000000), sdk.NewDec(1000000000)},
			expPanic:       false,
		},
		"get first pool asset, uneven two-asset pool with default scaling factors": {
			denoms:         []string{"foo"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []sdk.Dec{sdk.NewDec(2000000000)},
			expPanic:       false,
		},
		"get second pool asset, uneven two-asset pool with default scaling factors": {
			denoms:         []string{"bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []sdk.Dec{sdk.NewDec(1000000000)},
			expPanic:       false,
		},
		"get both pool assets, even two-asset pool with even scaling factors greater than 1": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10, 10},
			expReserves:    []sdk.Dec{sdk.NewDec(100000000), sdk.NewDec(100000000)},
			expPanic:       false,
		},
		"get both pool assets, even two-asset pool with uneven scaling factors greater than 1": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expReserves:    []sdk.Dec{sdk.NewDec(2000000000 / 5), sdk.NewDec(1000000000 / 10)},
			expPanic:       false,
		},
		"get first pool asset, even two-asset pool with uneven scaling factors greater than 1": {
			denoms:         []string{"foo"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expReserves:    []sdk.Dec{sdk.NewDec(2000000000 / 5)},
			expPanic:       false,
		},
		"get second pool asset, even two-asset pool with uneven scaling factors greater than 1": {
			denoms:         []string{"bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expReserves:    []sdk.Dec{sdk.NewDec(1000000000 / 10)},
			expPanic:       false,
		},
		"get both pool assets, even two-asset pool with even, massive scaling factors greater than 1": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000000000, 10000000000},
			expReserves:    []sdk.Dec{sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1)},
			expPanic:       false,
		},
		"max scaling factors": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expReserves:    []sdk.Dec{sdk.NewDec(1000000000).QuoInt64(int64(1 << 62)), sdk.NewDec(1000000000).QuoInt64(int64(1 << 62))},
			expPanic:       false,
		},
		"pass in no denoms": {
			denoms:         []string{},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []sdk.Dec{},
			expPanic:       false,
		},
		"zero scaling factor": {
			denoms:         []string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{0, 1},
			expPanic:       true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// we create the pool directly to bypass checks in NewStableswapPool()
				p := Pool{
					Address:            types.NewPoolAddress(defaultPoolId).String(),
					Id:                 defaultPoolId,
					PoolParams:         defaultStableswapPoolParams,
					TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(defaultPoolId), types.InitPoolSharesSupply),
					PoolLiquidity:      tc.poolAssets,
					ScalingFactor:      tc.scalingFactors,
					FuturePoolGovernor: defaultFutureGovernor,
				}

				reserves, err := p.getScaledPoolAmts(tc.denoms...)

				require.NoError(t, err, "test: %s", name)
				require.Equal(t, tc.expReserves, reserves)
			}

			osmoassert.ConditionalPanic(t, tc.expPanic, sut)
		})
	}
}

func TestGetDescaledPoolAmts(t *testing.T) {
	tests := map[string]struct {
		denom          string
		amount         osmomath.BigDec
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expResult      osmomath.BigDec
		expPanic       bool
	}{
		"pass in no denoms": {
			denom:          "",
			amount:         osmomath.ZeroDec(),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.ZeroDec(),
			expPanic:       false,
		},
		// sanity checks, default scaling factors
		"get exact supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(1000000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(1000000000),
			expPanic:       false,
		},
		"get less than supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(500000000),
			expPanic:       false,
		},
		"get more than supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(10000000000000),
			expPanic:       false,
		},

		// uneven pools
		"get exact supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(2000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(2000000000),
			expPanic:       false,
		},
		"get less than supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(500000000),
			expPanic:       false,
		},
		"get more than supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(10000000000000),
			expPanic:       false,
		},
		"get exact supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(1000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(1000000000),
			expPanic:       false,
		},
		"get less than supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(500000000),
			expPanic:       false,
		},
		"get more than supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      osmomath.NewBigDec(10000000000000),
			expPanic:       false,
		},

		// uneven scaling factors (note: denoms are ordered lexicographically, not by pool asset input)
		"get exact supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(2000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      osmomath.NewBigDec(2000000000 * 5),
			expPanic:       false,
		},
		"get less than supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      osmomath.NewBigDec(500000000 * 5),
			expPanic:       false,
		},
		"get more than supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      osmomath.NewBigDec(10000000000000 * 5),
			expPanic:       false,
		},
		"get exact supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(2000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      osmomath.NewBigDec(2000000000 * 10),
			expPanic:       false,
		},
		"get less than supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      osmomath.NewBigDec(500000000 * 10),
			expPanic:       false,
		},
		"get more than supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      osmomath.NewBigDec(10000000000000 * 10),
			expPanic:       false,
		},

		// panic catching
		"scaled asset overflows": {
			denom:          "foo",
			amount:         overflowDec,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expPanic:       true,
		},
		"descaled asset overflows": {
			denom: "foo",
			// 2^1000, should not overflow until descaled
			amount:         osmomath.NewDecFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1000), nil), big.NewInt(1))),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expPanic:       true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// we create the pool directly to bypass checks in NewStableswapPool()
				p := Pool{
					Address:            types.NewPoolAddress(defaultPoolId).String(),
					Id:                 defaultPoolId,
					PoolParams:         defaultStableswapPoolParams,
					TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(defaultPoolId), types.InitPoolSharesSupply),
					PoolLiquidity:      tc.poolAssets,
					ScalingFactor:      tc.scalingFactors,
					FuturePoolGovernor: defaultFutureGovernor,
				}

				result := p.getDescaledPoolAmt(tc.denom, tc.amount)
				require.Equal(t, tc.expResult, result)
			}

			osmoassert.ConditionalPanic(t, tc.expPanic, sut)
		})
	}
}

func TestScaledInput(t *testing.T) {
	tests := map[string]struct {
		input          sdk.Coins
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expOutput      []sdk.DecCoin
		expPanic       bool
	}{
		"even two-asset pool with default scaling factors": {
			input: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(100)),
				sdk.NewCoin("bar", sdk.NewInt(100)),
			),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expOutput:      []sdk.DecCoin{{"bar", sdk.NewDec(100)}, {"foo", sdk.NewDec(100)}},
			expPanic:       false,
		},
		"uneven two-asset pool with default scaling factors": {
			input: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(200)),
				sdk.NewCoin("bar", sdk.NewInt(100)),
			),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expOutput:      []sdk.DecCoin{{"bar", sdk.NewDec(100)}, {"foo", sdk.NewDec(200)}},
			expPanic:       false,
		},
		"even two-asset pool with even scaling factors greater than 1": {
			input: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(100)),
				sdk.NewCoin("bar", sdk.NewInt(100)),
			),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10, 10},
			expOutput:      []sdk.DecCoin{{"bar", sdk.NewDec(10)}, {"foo", sdk.NewDec(10)}},
			expPanic:       false,
		},
		"even two-asset pool with uneven scaling factors greater than 1": {
			input: sdk.NewCoins(
				sdk.NewCoin("bar", sdk.NewInt(100)),
				sdk.NewCoin("foo", sdk.NewInt(100)),
			),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expOutput:      []sdk.DecCoin{{"bar", sdk.NewDec(10)}, {"foo", sdk.NewDec(20)}},
			expPanic:       false,
		},
		"even two-asset pool with even, massive scaling factors greater than 1": {
			input: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(100)),
				sdk.NewCoin("bar", sdk.NewInt(100)),
			),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000000000, 10000000000},
			expOutput:      []sdk.DecCoin{{"bar", sdk.NewDec(100).Quo(sdk.NewDec(10000000000))}, {"foo", sdk.NewDec(100).Quo(sdk.NewDec(10000000000))}},
			expPanic:       false,
		},
		"five asset pool, scaling factors = 1, single-asset input": {
			input:          sdk.NewCoins(sdk.NewCoin("asset/c", sdk.NewInt(100))),
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 1, 1, 1, 1},
			expOutput:      []sdk.DecCoin{{"asset/c", sdk.NewDec(100)}},
			expPanic:       false,
		},
		"five asset pool, scaling factors = 1, all-asset input": {
			input: sdk.NewCoins(
				sdk.NewCoin("asset/a", sdk.NewInt(100)),
				sdk.NewCoin("asset/b", sdk.NewInt(100)),
				sdk.NewCoin("asset/c", sdk.NewInt(100)),
				sdk.NewCoin("asset/d", sdk.NewInt(100)),
				sdk.NewCoin("asset/e", sdk.NewInt(100)),
			),
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 1, 1, 1, 1},
			expOutput: []sdk.DecCoin{
				{"asset/a", sdk.NewDec(100)},
				{"asset/b", sdk.NewDec(100)},
				{"asset/c", sdk.NewDec(100)},
				{"asset/d", sdk.NewDec(100)},
				{"asset/e", sdk.NewDec(100)},
			},
			expPanic: false,
		},
		"five asset pool, scaling factors = 1,2,3,4,5, single-asset in": {
			input:          sdk.NewCoins(sdk.NewCoin("asset/d", sdk.NewInt(100))),
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expOutput:      []sdk.DecCoin{{"asset/d", sdk.NewDec(100).Quo(sdk.NewDec(4))}},
			expPanic:       false,
		},
		"five asset pool, scaling factors = 1,2,3,4,5, all-asset in": {
			input: sdk.NewCoins(
				sdk.NewCoin("asset/a", sdk.NewInt(100)),
				sdk.NewCoin("asset/b", sdk.NewInt(100)),
				sdk.NewCoin("asset/c", sdk.NewInt(100)),
				sdk.NewCoin("asset/d", sdk.NewInt(100)),
				sdk.NewCoin("asset/e", sdk.NewInt(100)),
			),
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expOutput: []sdk.DecCoin{
				{"asset/a", sdk.NewDec(100)},
				{"asset/b", sdk.NewDec(100).Quo(sdk.NewDec(2))},
				{"asset/c", sdk.NewDec(100).Quo(sdk.NewDec(3))},
				{"asset/d", sdk.NewDec(100).Quo(sdk.NewDec(4))},
				{"asset/e", sdk.NewDec(100).Quo(sdk.NewDec(5))},
			},
			expPanic: false,
		},
		"max scaling factors on min token inputs": {
			input: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(1)),
				sdk.NewCoin("bar", sdk.NewInt(1)),
			),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expOutput: []sdk.DecCoin{
				{"bar", sdk.NewDec(1).QuoInt64(int64(1 << 62))},
				{"foo", sdk.NewDec(1).QuoInt64(int64(1 << 62))},
			},
			expPanic: false,
		},
		"zero scaling factor": {
			input: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(100)),
				sdk.NewCoin("bar", sdk.NewInt(100)),
			),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{0, 1},
			expPanic:       true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				// we create the pool directly to bypass checks in NewStableswapPool()
				p := Pool{
					Address:            types.NewPoolAddress(defaultPoolId).String(),
					Id:                 defaultPoolId,
					PoolParams:         defaultStableswapPoolParams,
					TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(defaultPoolId), types.InitPoolSharesSupply),
					PoolLiquidity:      tc.poolAssets,
					ScalingFactor:      tc.scalingFactors,
					FuturePoolGovernor: defaultFutureGovernor,
				}

				reserves, err := p.scaledInput(tc.input)

				require.NoError(t, err, "test: %s", name)
				require.Equal(t, tc.expOutput, reserves)
			}

			osmoassert.ConditionalPanic(t, tc.expPanic, sut)
		})
	}
}
