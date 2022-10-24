//nolint:composites
package stableswap

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/internal/cfmm_common"
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
	defaultTwoAssetScalingFactors   = []uint64{1, 1}
	defaultThreeAssetScalingFactors = []uint64{1, 1, 1}
	defaultFiveAssetScalingFactors  = []uint64{1, 1, 1, 1, 1}
	defaultFutureGovernor           = ""

	twoEvenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("foo", 1000000000),
		sdk.NewInt64Coin("bar", 1000000000),
	)
	twoUnevenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("foo", 2000000000),
		sdk.NewInt64Coin("bar", 1000000000),
	)
	threeEvenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("asset/a", 1000000),
		sdk.NewInt64Coin("asset/b", 1000000),
		sdk.NewInt64Coin("asset/c", 1000000),
	)
	threeUnevenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("asset/a", 1000000),
		sdk.NewInt64Coin("asset/b", 2000000),
		sdk.NewInt64Coin("asset/c", 3000000),
	)
	fiveEvenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("asset/a", 1000000000),
		sdk.NewInt64Coin("asset/b", 1000000000),
		sdk.NewInt64Coin("asset/c", 1000000000),
		sdk.NewInt64Coin("asset/d", 1000000000),
		sdk.NewInt64Coin("asset/e", 1000000000),
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
				baseEvenAmt.MulInt64(3),
				baseEvenAmt.MulInt64(4),
				baseEvenAmt,
				baseEvenAmt.MulInt64(2),
				baseEvenAmt.MulInt64(5),
			},
		},
		"five asset pool, scaling factors = 1,2,3,4,5": {
			denoms:         [2]string{"asset/a", "asset/e"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expReserves: []osmomath.BigDec{
				baseEvenAmt,
				baseEvenAmt,
				baseEvenAmt,
				baseEvenAmt,
				baseEvenAmt,
			},
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

func TestSwapOutAmtGivenIn(t *testing.T) {
	tests := map[string]struct {
		poolAssets            sdk.Coins
		scalingFactors        []uint64
		tokenIn               sdk.Coins
		expectedTokenOut      sdk.Coin
		expectedPoolLiquidity sdk.Coins
		swapFee               sdk.Dec
		expError              bool
	}{
		"even pool basic trade": {
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			tokenIn:        sdk.NewCoins(sdk.NewInt64Coin("foo", 100)),
			// we expect at least a 1 token difference since output is truncated
			expectedTokenOut:      sdk.NewInt64Coin("bar", 99),
			expectedPoolLiquidity: twoEvenStablePoolAssets.Add(sdk.NewInt64Coin("foo", 100)).Sub(sdk.NewCoins(sdk.NewInt64Coin("bar", 99))),
			swapFee:               sdk.ZeroDec(),
			expError:              false,
		},
		"trade hits max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_998),
				sdk.NewInt64Coin("bar", 9_999_999_999),
			),
			scalingFactors:   defaultTwoAssetScalingFactors,
			tokenIn:          sdk.NewCoins(sdk.NewInt64Coin("foo", 1)),
			expectedTokenOut: sdk.NewInt64Coin("bar", 1),
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_999),
				sdk.NewInt64Coin("bar", 9_999_999_998),
			),
			swapFee:  sdk.ZeroDec(),
			expError: false,
		},
		"trade exceeds max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 10_000_000_000),
				sdk.NewInt64Coin("bar", 10_000_000_000),
			),
			scalingFactors:   defaultTwoAssetScalingFactors,
			tokenIn:          sdk.NewCoins(sdk.NewInt64Coin("foo", 1)),
			expectedTokenOut: sdk.Coin{},
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 10_000_000_000),
				sdk.NewInt64Coin("bar", 10_000_000_000),
			),
			swapFee:  sdk.ZeroDec(),
			expError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			tokenOut, err := p.SwapOutAmtGivenIn(ctx, tc.tokenIn, tc.expectedTokenOut.Denom, tc.swapFee)
			if !tc.expError {
				require.Equal(t, tc.expectedTokenOut, tokenOut)
				require.Equal(t, tc.expectedPoolLiquidity, p.PoolLiquidity)
			}
			osmoassert.ConditionalError(t, tc.expError, err)
		})
	}
}

func TestSwapInAmtGivenOut(t *testing.T) {
	tests := map[string]struct {
		poolAssets            sdk.Coins
		scalingFactors        []uint64
		tokenOut              sdk.Coins
		expectedTokenIn       sdk.Coin
		expectedPoolLiquidity sdk.Coins
		swapFee               sdk.Dec
		expError              bool
	}{
		"even pool basic trade": {
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			tokenOut:       sdk.NewCoins(sdk.NewInt64Coin("bar", 99)),
			// we expect at least a 1 token difference from our true expected output since it is truncated
			expectedTokenIn:       sdk.NewInt64Coin("foo", 99),
			expectedPoolLiquidity: twoEvenStablePoolAssets.Add(sdk.NewInt64Coin("foo", 99)).Sub(sdk.NewCoins(sdk.NewInt64Coin("bar", 99))),
			swapFee:               sdk.ZeroDec(),
			expError:              false,
		},
		"trade hits max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_998),
				sdk.NewInt64Coin("bar", 9_999_999_999),
			),
			scalingFactors:  defaultTwoAssetScalingFactors,
			tokenOut:        sdk.NewCoins(sdk.NewInt64Coin("bar", 1)),
			expectedTokenIn: sdk.NewInt64Coin("foo", 1),
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_999),
				sdk.NewInt64Coin("bar", 9_999_999_998),
			),
			swapFee:  sdk.ZeroDec(),
			expError: false,
		},
		"trade exceeds max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 10_000_000_000),
				sdk.NewInt64Coin("bar", 10_000_000_000),
			),
			scalingFactors:  defaultTwoAssetScalingFactors,
			tokenOut:        sdk.NewCoins(sdk.NewInt64Coin("bar", 1)),
			expectedTokenIn: sdk.Coin{},
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 10_000_000_000),
				sdk.NewInt64Coin("bar", 10_000_000_000),
			),
			swapFee:  sdk.ZeroDec(),
			expError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			tokenIn, err := p.SwapInAmtGivenOut(ctx, tc.tokenOut, tc.expectedTokenIn.Denom, tc.swapFee)
			if !tc.expError {
				require.Equal(t, tc.expectedTokenIn, tokenIn)
				require.Equal(t, tc.expectedPoolLiquidity, p.PoolLiquidity)
			}
			osmoassert.ConditionalError(t, tc.expError, err)
		})
	}
}

func TestInverseJoinPoolExitPool(t *testing.T) {
	hundredFoo := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(100)))
	thousandAssetA := sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(1000)))
	tenPercentOfTwoPoolRaw := int64(1000000000 / 10)
	tenPercentOfTwoPoolCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(int64(1000000000/10))), sdk.NewCoin("bar", sdk.NewInt(int64(1000000000/10))))
	type testcase struct {
		tokensIn           sdk.Coins
		poolAssets         sdk.Coins
		unevenJoinedTokens sdk.Coins
		scalingFactors     []uint64
		swapFee            sdk.Dec
		expectPass         bool
	}

	tests := map[string]testcase{
		"[single asset join] even two asset pool, no swap fee": {
			tokensIn:       hundredFoo,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
		"[single asset join] uneven two asset pool, no swap fee": {
			tokensIn:       hundredFoo,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
		"[single asset join] even 3-asset pool, no swap fee": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
		"[single asset join] uneven 3-asset pool, no swap fee": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
		"[single asset join] even two asset pool, default swap fee": {
			tokensIn:       hundredFoo,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectPass:     true,
		},
		"[single asset join] uneven two asset pool, default swap fee": {
			tokensIn:       hundredFoo,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectPass:     true,
		},
		"[single asset join] even 3-asset pool, default swap fee": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectPass:     true,
		},
		"[single asset join] uneven 3-asset pool, default swap fee": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectPass:     true,
		},
		"[single asset join] even 3-asset pool, 0.03 swap fee": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.MustNewDecFromStr("0.03"),
			expectPass:     true,
		},
		"[single asset join] uneven 3-asset pool, 0.03 swap fee": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			swapFee:        sdk.MustNewDecFromStr("0.03"),
			expectPass:     true,
		},

		"[all asset join] even two asset pool, same tokenIn ratio": {
			tokensIn:       tenPercentOfTwoPoolCoins,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
		"[all asset join] even two asset pool, different tokenIn ratio with pool": {
			tokensIn:       sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPoolRaw)), sdk.NewCoin("bar", sdk.NewInt(10+tenPercentOfTwoPoolRaw))),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
		"[all asset join] even two asset pool, different tokenIn ratio with pool, nonzero swap fee": {
			tokensIn:       sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPoolRaw)), sdk.NewCoin("bar", sdk.NewInt(10+tenPercentOfTwoPoolRaw))),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        defaultSwapFee,
			expectPass:     true,
		},
		"[all asset join] even two asset pool, no tokens in": {
			tokensIn:       sdk.NewCoins(),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			swapFee:        sdk.ZeroDec(),
			expectPass:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			// we join then exit the pool
			shares, err := p.JoinPool(ctx, tc.tokensIn, tc.swapFee)
			tokenOut, err := p.ExitPool(ctx, shares, defaultExitFee)

			// if single asset join, we swap output tokens to input denom to test the full inverse relationship
			if len(tc.tokensIn) == 1 {
				tokenOutAmt, err := cfmm_common.SwapAllCoinsToSingleAsset(&p, ctx, tokenOut, tc.tokensIn[0].Denom)
				require.NoError(t, err)
				tokenOut = sdk.NewCoins(sdk.NewCoin(tc.tokensIn[0].Denom, tokenOutAmt))
			}

			// if single asset join, we expect output token swapped into the input denom to be input minus swap fee
			var expectedTokenOut sdk.Coins
			if len(tc.tokensIn) == 1 {
				expectedAmt := (tc.tokensIn[0].Amount.ToDec().Mul(sdk.OneDec().Sub(tc.swapFee))).TruncateInt()
				expectedTokenOut = sdk.NewCoins(sdk.NewCoin(tc.tokensIn[0].Denom, expectedAmt))
			} else {
				expectedTokenOut = tc.tokensIn
			}

			if tc.expectPass {
				finalPoolLiquidity := p.GetTotalPoolLiquidity(ctx)
				require.True(t, tokenOut.IsAllLTE(expectedTokenOut))
				require.True(t, finalPoolLiquidity.IsAllGTE(tc.poolAssets))
			}
			osmoassert.ConditionalError(t, !tc.expectPass, err)
		})
	}
}
