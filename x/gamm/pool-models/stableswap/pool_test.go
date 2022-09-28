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
