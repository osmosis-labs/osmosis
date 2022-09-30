//nolint:composites
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

func TestScaledSortedPoolReserves(t *testing.T) {
	baseEvenAmt := osmomath.NewBigDec(1000000000)
	tests := map[string]struct {
		denoms         [2]string
		roundMode      osmomath.RoundingDirection
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expReserves    []osmomath.BigDec
		expError       bool
	}{
		// sanity checks, default scaling factors
		"even two-asset pool with default scaling factors": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []osmomath.BigDec{baseEvenAmt, baseEvenAmt},
		},
		"uneven two-asset pool with default scaling factors": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []osmomath.BigDec{baseEvenAmt.MulInt64(2), baseEvenAmt},
		},
		"even two-asset pool with even scaling factors greater than 1": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10, 10},
			expReserves:    []osmomath.BigDec{baseEvenAmt.QuoInt64(10), baseEvenAmt.QuoInt64(10)},
		},
		"even two-asset pool with uneven scaling factors greater than 1": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expReserves: []osmomath.BigDec{
				osmomath.NewBigDec(2000000000 / 5), osmomath.NewBigDec(1000000000 / 10)},
		},
		"even two-asset pool with even, massive scaling factors greater than 1": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000000000, 10000000000},
			expReserves:    []osmomath.BigDec{osmomath.NewDecWithPrec(1, 1), osmomath.NewDecWithPrec(1, 1)},
		},
		"five asset pool, scaling factors = 1": {
			denoms:         [2]string{"asset/c", "asset/d"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 1, 1, 1, 1},
			expReserves: []osmomath.BigDec{
				baseEvenAmt.MulInt64(3),  // {"asset/c", baseEvenAmt.MulInt64(3)},
				baseEvenAmt.MulInt64(4),  // {"asset/d", baseEvenAmt.MulInt64(4)},
				baseEvenAmt,              // {"asset/a", baseEvenAmt},
				baseEvenAmt.MulInt64(2),  // {"asset/b", baseEvenAmt.MulInt64(2)},
				baseEvenAmt.MulInt64(5)}, // {"asset/e", baseEvenAmt.MulInt64(5)}},
		},
		"five asset pool, scaling factors = 1,2,3,4,5": {
			denoms:         [2]string{"asset/a", "asset/e"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expReserves: []osmomath.BigDec{
				baseEvenAmt,  // {"asset/a", baseEvenAmt},
				baseEvenAmt,  // {"asset/e", baseEvenAmt},
				baseEvenAmt,  // {"asset/b", baseEvenAmt},
				baseEvenAmt,  // {"asset/c", baseEvenAmt},
				baseEvenAmt}, // {"asset/d", baseEvenAmt}},
		},
		"max scaling factors": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expReserves: []osmomath.BigDec{
				osmomath.NewBigDec(1000000000).QuoInt64(int64(1 << 62)),
				osmomath.NewBigDec(1000000000).QuoInt64(int64(1 << 62))},
		},
		"zero scaling factor": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{0, 1},
			expError:       true,
		},
	}
	// TODO: Add for loop for trying to re-order test cases

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.roundMode == 0 {
				tc.roundMode = osmomath.RoundBankers
			}
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

			reserves, err := p.scaledSortedPoolReserves(tc.denoms[0], tc.denoms[1], tc.roundMode)
			if !tc.expError {
				require.Equal(t, tc.expReserves, reserves)
			}
			osmoassert.ConditionalError(t, tc.expError, err)
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
