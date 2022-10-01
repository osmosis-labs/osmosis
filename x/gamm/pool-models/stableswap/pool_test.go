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

// we create a pool struct directly to bypass checks in NewStableswapPool()
func poolStructFromAssets(assets sdk.Coins, scalingFactors []uint64) Pool {
	p := Pool{
		Address:            types.NewPoolAddress(defaultPoolId).String(),
		Id:                 defaultPoolId,
		PoolParams:         defaultStableswapPoolParams,
		TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(defaultPoolId), types.InitPoolSharesSupply),
		PoolLiquidity:      assets,
		ScalingFactor:      scalingFactors,
		FuturePoolGovernor: defaultFutureGovernor,
	}
	return p
}

func TestReorderReservesAndScalingFactors(t *testing.T) {
	tests := map[string]struct {
		denoms                [2]string
		poolAssets            sdk.Coins
		scalingFactors        []uint64
		reordedReserves       []sdk.Coin
		reordedScalingFactors []uint64
		expError              bool
	}{
		"two of 5 assets in pool": {
			denoms:         [2]string{"asset/c", "asset/b"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			reordedReserves: []sdk.Coin{
				sdk.NewInt64Coin("asset/c", 3000000000),
				sdk.NewInt64Coin("asset/b", 2000000000),
				sdk.NewInt64Coin("asset/a", 1000000000),
				sdk.NewInt64Coin("asset/d", 4000000000),
				sdk.NewInt64Coin("asset/e", 5000000000),
			},
			reordedScalingFactors: []uint64{3, 2, 1, 4, 5},
		},
		"two of 5 assets in pool v2": {
			denoms:         [2]string{"asset/e", "asset/b"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			reordedReserves: []sdk.Coin{
				sdk.NewInt64Coin("asset/e", 5000000000),
				sdk.NewInt64Coin("asset/b", 2000000000),
				sdk.NewInt64Coin("asset/a", 1000000000),
				sdk.NewInt64Coin("asset/c", 3000000000),
				sdk.NewInt64Coin("asset/d", 4000000000),
			},
			reordedScalingFactors: []uint64{5, 2, 1, 3, 4},
		},
		"asset 1 doesn't exist": {
			denoms:         [2]string{"foo", "asset/b"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expError:       true,
		},
		"asset 2 doesn't exist": {
			denoms:         [2]string{"asset/a", "foo"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expError:       true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			reserves, factors, err := p.reorderReservesAndScalingFactors(tc.denoms[0], tc.denoms[1])
			if !tc.expError {
				require.Equal(t, tc.reordedReserves, reserves)
				require.Equal(t, tc.reordedScalingFactors, factors)
			}
			osmoassert.ConditionalError(t, tc.expError, err)
		})
	}
}

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
				osmomath.NewBigDec(2000000000 / 5), osmomath.NewBigDec(1000000000 / 10),
			},
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
				baseEvenAmt.MulInt64(3), // {"asset/c", baseEvenAmt.MulInt64(3)},
				baseEvenAmt.MulInt64(4), // {"asset/d", baseEvenAmt.MulInt64(4)},
				baseEvenAmt,             // {"asset/a", baseEvenAmt},
				baseEvenAmt.MulInt64(2), // {"asset/b", baseEvenAmt.MulInt64(2)},
				baseEvenAmt.MulInt64(5),
			}, // {"asset/e", baseEvenAmt.MulInt64(5)}},
		},
		"five asset pool, scaling factors = 1,2,3,4,5": {
			denoms:         [2]string{"asset/a", "asset/e"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expReserves: []osmomath.BigDec{
				baseEvenAmt, // {"asset/a", baseEvenAmt},
				baseEvenAmt, // {"asset/e", baseEvenAmt},
				baseEvenAmt, // {"asset/b", baseEvenAmt},
				baseEvenAmt, // {"asset/c", baseEvenAmt},
				baseEvenAmt,
			}, // {"asset/d", baseEvenAmt}},
		},
		"max scaling factors": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expReserves: []osmomath.BigDec{
				osmomath.NewBigDec(1000000000).QuoInt64(int64(1 << 62)),
				osmomath.NewBigDec(1000000000).QuoInt64(int64(1 << 62)),
			},
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
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

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
		expResult      sdk.Dec
		expPanic       bool
	}{
		"pass in no denoms": {
			denom:          "",
			amount:         osmomath.ZeroDec(),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.ZeroDec(),
		},
		// sanity checks, default scaling factors
		"get exact supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(1000000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(1000000000),
		},
		"get less than supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(500000000),
		},
		"get more than supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(10000000000000),
		},

		// uneven pools
		"get exact supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(2000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(2000000000),
		},
		"get less than supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(500000000),
		},
		"get more than supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(10000000000000),
		},
		"get exact supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(1000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(1000000000),
		},
		"get less than supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(500000000),
		},
		"get more than supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(10000000000000),
		},

		// uneven scaling factors (note: denoms are ordered lexicographically, not by pool asset input)
		"get exact supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(2000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(2000000000 * 5),
		},
		"get less than supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(500000000 * 5),
		},
		"get more than supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(10000000000000 * 5),
		},
		"get exact supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(2000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(2000000000 * 10),
		},
		"get less than supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(500000000 * 10),
		},
		"get more than supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(10000000000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(10000000000000 * 10),
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

func TestScaleCoin(t *testing.T) {
	tests := map[string]struct {
		input          sdk.Coin
		rounding       osmomath.RoundingDirection
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expOutput      osmomath.BigDec
		expError       bool
	}{
		"even two-asset pool with default scaling factors": {
			input:          sdk.NewCoin("bar", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expOutput:      osmomath.NewBigDec(100),
			expError:       false,
		},
		"uneven two-asset pool with default scaling factors": {
			input:          sdk.NewCoin("foo", sdk.NewInt(200)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expOutput:      osmomath.NewBigDec(200),
			expError:       false,
		},
		"even two-asset pool with uneven scaling factors greater than 1": {
			input:          sdk.NewCoin("bar", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expOutput:      osmomath.NewBigDec(10),
			expError:       false,
		},
		"even two-asset pool with even, massive scaling factors greater than 1": {
			input:          sdk.NewCoin("foo", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000000000, 10_000_000_000},
			expOutput:      osmomath.NewDecWithPrec(100, 10),
			expError:       false,
		},
		"five asset pool scaling factors = 1": {
			input:          sdk.NewCoin("asset/c", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 1, 1, 1, 1},
			expOutput:      osmomath.NewBigDec(100),
			expError:       false,
		},
		"five asset pool scaling factors = 1,2,3,4,5": {
			input:          sdk.NewCoin("asset/d", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expOutput:      osmomath.NewBigDec(25),
			expError:       false,
		},
		"max scaling factors on small token inputs": {
			input:          sdk.NewCoin("foo", sdk.NewInt(10)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62), (1 << 62)},
			expOutput:      osmomath.NewBigDec(10).QuoInt64(1 << 62),
			expError:       false,
		},
		"zero scaling factor": {
			input:          sdk.NewCoin("bar", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{0, 1},
			expError:       true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

			scaledInput, err := p.scaleCoin(tc.input, tc.rounding)

			if !tc.expError {
				require.NoError(t, err, "test: %s", name)
				require.Equal(t, tc.expOutput, scaledInput)
			}

			osmoassert.ConditionalError(t, tc.expError, err)
		})
	}
}
