//nolint:composites
package stableswap

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/internal/cfmm_common"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var (
	defaultSpreadFactor         = sdk.MustNewDecFromStr("0.025")
	defaultExitFee              = sdk.ZeroDec()
	defaultPoolId               = uint64(1)
	defaultStableswapPoolParams = PoolParams{
		SwapFee: defaultSpreadFactor,
		ExitFee: defaultExitFee,
	}
	defaultTwoAssetScalingFactors   = []uint64{1, 1}
	defaultThreeAssetScalingFactors = []uint64{1, 1, 1}
	defaultFiveAssetScalingFactors  = []uint64{1, 1, 1, 1, 1}
	defaultFutureGovernor           = ""

	twoEvenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("bar", 1000000000),
		sdk.NewInt64Coin("foo", 1000000000),
	)
	twoUnevenStablePoolAssets = sdk.NewCoins(
		sdk.NewInt64Coin("bar", 1000000000),
		sdk.NewInt64Coin("foo", 2000000000),
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
	scalingFactors, _ = applyScalingFactorMultiplier(scalingFactors)

	p := Pool{
		Address:            poolmanagertypes.NewPoolAddress(defaultPoolId).String(),
		Id:                 defaultPoolId,
		PoolParams:         defaultStableswapPoolParams,
		TotalShares:        sdk.NewCoin(types.GetPoolShareDenom(defaultPoolId), types.InitPoolSharesSupply),
		PoolLiquidity:      assets,
		ScalingFactors:     scalingFactors,
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
			reordedScalingFactors: []uint64{3 * types.ScalingFactorMultiplier, 2 * types.ScalingFactorMultiplier, 1 * types.ScalingFactorMultiplier, 4 * types.ScalingFactorMultiplier, 5 * types.ScalingFactorMultiplier},
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
			reordedScalingFactors: []uint64{5 * types.ScalingFactorMultiplier, 2 * types.ScalingFactorMultiplier, 1 * types.ScalingFactorMultiplier, 3 * types.ScalingFactorMultiplier, 4 * types.ScalingFactorMultiplier},
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
	bigDecScalingMultiplier := osmomath.NewBigDec(types.ScalingFactorMultiplier)
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
			expReserves:    []osmomath.BigDec{baseEvenAmt.Quo(bigDecScalingMultiplier), baseEvenAmt.Quo(bigDecScalingMultiplier)},
		},
		"uneven two-asset pool with default scaling factors": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expReserves:    []osmomath.BigDec{baseEvenAmt.MulInt64(2).Quo(bigDecScalingMultiplier), baseEvenAmt.Quo(bigDecScalingMultiplier)},
		},
		"even two-asset pool with even scaling factors greater than 1": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10, 10},
			expReserves:    []osmomath.BigDec{(baseEvenAmt.Quo(bigDecScalingMultiplier)).QuoInt64(10), (baseEvenAmt.Quo(bigDecScalingMultiplier)).QuoInt64(10)},
		},
		"even two-asset pool with uneven scaling factors greater than 1": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expReserves: []osmomath.BigDec{
				osmomath.NewBigDec(2000000000 / 5).Quo(bigDecScalingMultiplier), osmomath.NewBigDec(1000000000 / 10).Quo(bigDecScalingMultiplier),
			},
		},
		"even two-asset pool with even, massive scaling factors greater than 1": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000000000, 10000000000},
			expReserves:    []osmomath.BigDec{osmomath.NewDecWithPrec(1, 1).Quo(bigDecScalingMultiplier), osmomath.NewDecWithPrec(1, 1).Quo(bigDecScalingMultiplier)},
		},
		"five asset pool, scaling factors = 1": {
			denoms:         [2]string{"asset/c", "asset/d"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 1, 1, 1, 1},
			expReserves: []osmomath.BigDec{
				baseEvenAmt.MulInt64(3).Quo(bigDecScalingMultiplier),
				baseEvenAmt.MulInt64(4).Quo(bigDecScalingMultiplier),
				baseEvenAmt.Quo(bigDecScalingMultiplier),
				baseEvenAmt.MulInt64(2).Quo(bigDecScalingMultiplier),
				baseEvenAmt.MulInt64(5).Quo(bigDecScalingMultiplier),
			},
		},
		"five asset pool, scaling factors = 1,2,3,4,5": {
			denoms:         [2]string{"asset/a", "asset/e"},
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expReserves: []osmomath.BigDec{
				baseEvenAmt.Quo(bigDecScalingMultiplier),
				baseEvenAmt.Quo(bigDecScalingMultiplier),
				baseEvenAmt.Quo(bigDecScalingMultiplier),
				baseEvenAmt.Quo(bigDecScalingMultiplier),
				baseEvenAmt.Quo(bigDecScalingMultiplier),
			},
		},
		"max scaling factors": {
			denoms:         [2]string{"foo", "bar"},
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62) / types.ScalingFactorMultiplier, (1 << 62) / types.ScalingFactorMultiplier},
			expReserves: []osmomath.BigDec{
				(osmomath.NewBigDec(1000000000).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier))).Quo(osmomath.NewBigDec(int64(1<<62) / types.ScalingFactorMultiplier)),
				(osmomath.NewBigDec(1000000000).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier))).Quo(osmomath.NewBigDec(int64(1<<62) / types.ScalingFactorMultiplier)),
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

func TestGetScalingFactorByDenom(t *testing.T) {
	tests := map[string]struct {
		denom          string
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expResult      uint64
	}{
		"pass in no denoms": {
			denom:          "",
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      0,
		},
		"get scaling factor for first asset (two-asset pool)": {
			denom:          "bar",
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{1, 2},
			expResult:      1,
		},
		"get scaling factor for second asset (two-asset pool)": {
			denom:          "foo",
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{1, 2},
			expResult:      2,
		},
		"get scaling factor for second asset (three-asset pool)": {
			denom:          "asset/b",
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3},
			expResult:      2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			factor := p.GetScalingFactorByDenom(tc.denom)
			require.Equal(t, tc.expResult, factor)
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
			amount:         osmomath.NewBigDec(100000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(100000000 * types.ScalingFactorMultiplier),
		},
		"get less than supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(500000000 * types.ScalingFactorMultiplier),
		},
		"get more than supply of one asset, even two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(100000000),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(100000000 * types.ScalingFactorMultiplier),
		},

		// uneven pools
		"get exact supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(200000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(200000000 * types.ScalingFactorMultiplier),
		},
		"get less than supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(500000000 * types.ScalingFactorMultiplier),
		},
		"get more than supply of first asset, uneven two-asset pool with default scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(100000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(100000000 * types.ScalingFactorMultiplier),
		},
		"get exact supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(100000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(100000000 * types.ScalingFactorMultiplier),
		},
		"get less than supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(500000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(500000000 * types.ScalingFactorMultiplier),
		},
		"get more than supply of second asset, uneven two-asset pool with default scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(100000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expResult:      sdk.NewDec(100000000 * types.ScalingFactorMultiplier),
		},

		// uneven scaling factors (note: denoms are ordered lexicographically, not by pool asset input)
		"get exact supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(20000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(20000000 * 5 * types.ScalingFactorMultiplier),
		},
		"get less than supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(50000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(50000000 * 5 * types.ScalingFactorMultiplier),
		},
		"get more than supply of first asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "foo",
			amount:         osmomath.NewBigDec(100000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(100000000 * 5 * types.ScalingFactorMultiplier),
		},
		"get exact supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(20000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(20000000 * 10 * types.ScalingFactorMultiplier),
		},
		"get less than supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(50000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(50000000 * 10 * types.ScalingFactorMultiplier),
		},
		"get more than supply of second asset, uneven two-asset pool with uneven scaling factors": {
			denom:          "bar",
			amount:         osmomath.NewBigDec(10000000),
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expResult:      sdk.NewDec(10000000 * 10 * types.ScalingFactorMultiplier),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			sut := func() {
				p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

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
			expOutput:      osmomath.NewBigDec(100).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier)),
			expError:       false,
		},
		"uneven two-asset pool with default scaling factors": {
			input:          sdk.NewCoin("foo", sdk.NewInt(200)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expOutput:      osmomath.NewBigDec(200).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier)),
			expError:       false,
		},
		"even two-asset pool with uneven scaling factors greater than 1": {
			input:          sdk.NewCoin("bar", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10, 5},
			expOutput:      osmomath.NewBigDec(10).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier)),
			expError:       false,
		},
		"even two-asset pool with even, massive scaling factors greater than 1": {
			input:          sdk.NewCoin("foo", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000000000, 10_000_000_000},
			expOutput:      osmomath.NewDecWithPrec(100, 10).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier)),
			expError:       false,
		},
		"five asset pool scaling factors = 1": {
			input:          sdk.NewCoin("asset/c", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 1, 1, 1, 1},
			expOutput:      osmomath.NewBigDec(100).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier)),
			expError:       false,
		},
		"five asset pool scaling factors = 1,2,3,4,5": {
			input:          sdk.NewCoin("asset/d", sdk.NewInt(100)),
			rounding:       osmomath.RoundDown,
			poolAssets:     fiveUnevenStablePoolAssets,
			scalingFactors: []uint64{1, 2, 3, 4, 5},
			expOutput:      osmomath.NewBigDec(25).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier)),
			expError:       false,
		},
		"max scaling factors on small token inputs": {
			input:          sdk.NewCoin("foo", sdk.NewInt(10)),
			rounding:       osmomath.RoundDown,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{(1 << 62) / types.ScalingFactorMultiplier, (1 << 62) / types.ScalingFactorMultiplier},
			expOutput:      (osmomath.NewBigDec(10).Quo(osmomath.NewBigDec(types.ScalingFactorMultiplier))).Quo(osmomath.NewBigDec((1 << 62) / types.ScalingFactorMultiplier)),
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
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			scaledInput, err := p.scaleCoin(tc.input, tc.rounding)

			if !tc.expError {
				require.NoError(t, err, "test: %s", name)
				require.Equal(t, tc.expOutput, scaledInput)
			}

			osmoassert.ConditionalError(t, tc.expError, err)
		})
	}
}

func TestCalcJoinPoolNoSwapShares(t *testing.T) {
	tenPercentOfTwoPool := int64(1000000000 / 10)
	tenPercentOfThreePool := int64(1000000 / 10)
	tests := map[string]struct {
		tokensIn        sdk.Coins
		poolAssets      sdk.Coins
		scalingFactors  []uint64
		expNumShare     sdk.Int
		expTokensJoined sdk.Coins
		expPoolAssets   sdk.Coins
		expectPass      bool
	}{
		"even two asset pool, same tokenIn ratio": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool))),
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool))),
			expPoolAssets:   twoEvenStablePoolAssets,
			expectPass:      true,
		},
		"even two asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool+1))),
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool))),
			expPoolAssets:   twoEvenStablePoolAssets,
			expectPass:      true,
		},
		"uneven two asset pool, same tokenIn ratio": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2*tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool))),
			poolAssets:      twoUnevenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2*tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool))),
			expPoolAssets:   twoUnevenStablePoolAssets,
			expectPass:      true,
		},
		"uneven two asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2*tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool+1))),
			poolAssets:      twoUnevenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(2*tenPercentOfTwoPool)), sdk.NewCoin("bar", sdk.NewInt(tenPercentOfTwoPool))),
			expPoolAssets:   twoUnevenStablePoolAssets,
			expectPass:      true,
		},
		"even three asset pool, same tokenIn ratio": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(tenPercentOfThreePool))),
			poolAssets:      threeEvenStablePoolAssets,
			scalingFactors:  defaultThreeAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(tenPercentOfThreePool))),
			expPoolAssets:   threeEvenStablePoolAssets,
			expectPass:      true,
		},
		"even three asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(tenPercentOfThreePool+1))),
			poolAssets:      threeEvenStablePoolAssets,
			scalingFactors:  defaultThreeAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(tenPercentOfThreePool))),
			expPoolAssets:   threeEvenStablePoolAssets,
			expectPass:      true,
		},
		"uneven three asset pool, same tokenIn ratio": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(2*tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(3*tenPercentOfThreePool))),
			poolAssets:      threeUnevenStablePoolAssets,
			scalingFactors:  defaultThreeAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(2*tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(3*tenPercentOfThreePool))),
			expPoolAssets:   threeUnevenStablePoolAssets,
			expectPass:      true,
		},
		"uneven three asset pool, different tokenIn ratio with pool": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(2*tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(3*tenPercentOfThreePool+1))),
			poolAssets:      threeUnevenStablePoolAssets,
			scalingFactors:  defaultThreeAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(2*tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(3*tenPercentOfThreePool))),
			expPoolAssets:   threeUnevenStablePoolAssets,
			expectPass:      true,
		},
		"uneven three asset pool, uneven scaling factors": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(2*tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(3*tenPercentOfThreePool))),
			poolAssets:      threeUnevenStablePoolAssets,
			scalingFactors:  []uint64{5, 9, 175},
			expNumShare:     sdk.NewIntFromUint64(10000000000000000000),
			expTokensJoined: sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(2*tenPercentOfThreePool)), sdk.NewCoin("asset/c", sdk.NewInt(3*tenPercentOfThreePool))),
			expPoolAssets:   threeUnevenStablePoolAssets,
			expectPass:      true,
		},

		// error catching
		"even two asset pool, no-swap join attempt with one asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPool))),
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   twoEvenStablePoolAssets,
			expectPass:      false,
		},
		"even two asset pool, no-swap join attempt with one valid and one invalid asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPool)), sdk.NewCoin("baz", sdk.NewInt(tenPercentOfTwoPool))),
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   twoEvenStablePoolAssets,
			expectPass:      false,
		},
		"even two asset pool, no-swap join attempt with two invalid assets": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("baz", sdk.NewInt(tenPercentOfTwoPool)), sdk.NewCoin("qux", sdk.NewInt(tenPercentOfTwoPool))),
			poolAssets:      twoEvenStablePoolAssets,
			scalingFactors:  defaultTwoAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   twoEvenStablePoolAssets,
			expectPass:      false,
		},
		"even three asset pool, no-swap join attempt with an invalid asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("asset/b", sdk.NewInt(tenPercentOfThreePool)), sdk.NewCoin("qux", sdk.NewInt(tenPercentOfThreePool))),
			poolAssets:      threeEvenStablePoolAssets,
			scalingFactors:  defaultThreeAssetScalingFactors,
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   threeEvenStablePoolAssets,
			expectPass:      false,
		},
		"single asset pool, no-swap join attempt with one asset": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(sdk.MaxSortableDec.TruncateInt64()))),
			poolAssets:      sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			scalingFactors:  []uint64{1},
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			expectPass:      false,
		},
		"attempt joining pool with no assets in it": {
			tokensIn:        sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1))),
			poolAssets:      sdk.Coins{},
			scalingFactors:  []uint64{},
			expNumShare:     sdk.NewIntFromUint64(0),
			expTokensJoined: sdk.Coins{},
			expPoolAssets:   sdk.Coins{},
			expectPass:      false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			pool := poolStructFromAssets(test.poolAssets, test.scalingFactors)
			numShare, tokensJoined, err := pool.CalcJoinPoolNoSwapShares(ctx, test.tokensIn, pool.GetSpreadFactor(ctx))

			if test.expectPass {
				require.NoError(t, err)
				require.Equal(t, test.expPoolAssets, pool.GetTotalPoolLiquidity(ctx))
				require.Equal(t, test.expNumShare, numShare)
				require.Equal(t, test.expTokensJoined, tokensJoined)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expPoolAssets, pool.GetTotalPoolLiquidity(ctx))
				require.Equal(t, test.expNumShare, numShare)
				require.Equal(t, test.expTokensJoined, tokensJoined)
			}
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
		spreadFactor          sdk.Dec
		expError              bool
	}{
		"even pool basic trade": {
			poolAssets:            twoEvenStablePoolAssets,
			scalingFactors:        defaultTwoAssetScalingFactors,
			tokenIn:               sdk.NewCoins(sdk.NewInt64Coin("foo", 100)),
			expectedTokenOut:      sdk.NewInt64Coin("bar", 99),
			expectedPoolLiquidity: twoEvenStablePoolAssets.Add(sdk.NewInt64Coin("foo", 100)).Sub(sdk.NewCoins(sdk.NewInt64Coin("bar", 99))),
			spreadFactor:          sdk.ZeroDec(),
			expError:              false,
		},
		"100:1 scaling factor ratio, even swap": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("bar", 1000000000),
				sdk.NewInt64Coin("foo", 10000000),
			),
			scalingFactors:   []uint64{100, 1},
			tokenIn:          sdk.NewCoins(sdk.NewInt64Coin("foo", 100)),
			expectedTokenOut: sdk.NewInt64Coin("bar", 9999),
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("bar", 1000000000).SubAmount(sdk.NewIntFromUint64(9999)),
				sdk.NewInt64Coin("foo", 10000000).AddAmount(sdk.NewIntFromUint64(100)),
			),
			spreadFactor: sdk.ZeroDec(),
			expError:     false,
		},
		// TODO: Add test cases here, where they're off 1-1 ratio
		// * (we just need to verify that the further off they are, further slippage is)
		// * Add test cases with non-zero spread factor.
		// looks like its really an error due to slippage at limit
		"trade hits max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_998),
				sdk.NewInt64Coin("bar", 9_999_999_998),
			),
			scalingFactors:   defaultTwoAssetScalingFactors,
			tokenIn:          sdk.NewCoins(sdk.NewInt64Coin("foo", 1)),
			expectedTokenOut: sdk.Coin{},
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_999),
				sdk.NewInt64Coin("bar", 9_999_999_997),
			),
			spreadFactor: sdk.ZeroDec(),
			expError:     true,
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
			spreadFactor: sdk.ZeroDec(),
			expError:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			tokenOut, err := p.SwapOutAmtGivenIn(ctx, tc.tokenIn, tc.expectedTokenOut.Denom, tc.spreadFactor)
			osmoassert.ConditionalError(t, tc.expError, err)
			if !tc.expError {
				require.Equal(t, tc.expectedTokenOut.Amount, tokenOut.Amount)
				require.True(t, p.PoolLiquidity.IsAllGTE(tc.expectedPoolLiquidity),
					"p.PoolLiquidity.IsAllGTE(tc.expectedPoolLiquidity) failed. Pool liq %v, expected %v",
					p.PoolLiquidity, tc.expectedPoolLiquidity)
			}
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
		spreadFactor          sdk.Dec
		expError              bool
	}{
		"even pool basic trade": {
			poolAssets:            twoEvenStablePoolAssets,
			scalingFactors:        defaultTwoAssetScalingFactors,
			tokenOut:              sdk.NewCoins(sdk.NewInt64Coin("bar", 100)),
			expectedTokenIn:       sdk.NewInt64Coin("foo", 100),
			expectedPoolLiquidity: twoEvenStablePoolAssets.Add(sdk.NewInt64Coin("foo", 100)).Sub(sdk.NewCoins(sdk.NewInt64Coin("bar", 100))),
			spreadFactor:          sdk.ZeroDec(),
			expError:              false,
		},
		"trade hits max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_997*types.ScalingFactorMultiplier),
				sdk.NewInt64Coin("bar", 9_999_999_997*types.ScalingFactorMultiplier),
			),
			scalingFactors:  defaultTwoAssetScalingFactors,
			tokenOut:        sdk.NewCoins(sdk.NewInt64Coin("bar", 1*types.ScalingFactorMultiplier)),
			expectedTokenIn: sdk.NewInt64Coin("foo", 1*types.ScalingFactorMultiplier),
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 9_999_999_998*types.ScalingFactorMultiplier),
				sdk.NewInt64Coin("bar", 9_999_999_996*types.ScalingFactorMultiplier),
			),
			spreadFactor: sdk.ZeroDec(),
			expError:     false,
		},
		"trade exceeds max pool capacity for asset": {
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 10_000_000_000*types.ScalingFactorMultiplier),
				sdk.NewInt64Coin("bar", 10_000_000_000*types.ScalingFactorMultiplier),
			),
			scalingFactors:  defaultTwoAssetScalingFactors,
			tokenOut:        sdk.NewCoins(sdk.NewInt64Coin("bar", 1)),
			expectedTokenIn: sdk.Coin{},
			expectedPoolLiquidity: sdk.NewCoins(
				sdk.NewInt64Coin("foo", 10_000_000_000*types.ScalingFactorMultiplier),
				sdk.NewInt64Coin("bar", 10_000_000_000*types.ScalingFactorMultiplier),
			),
			spreadFactor: sdk.ZeroDec(),
			expError:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			tokenIn, err := p.SwapInAmtGivenOut(ctx, tc.tokenOut, tc.expectedTokenIn.Denom, tc.spreadFactor)
			if !tc.expError {
				require.True(t, tokenIn.Amount.GTE(tc.expectedTokenIn.Amount))
				require.True(t, p.PoolLiquidity.IsAllGTE(tc.expectedPoolLiquidity))
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
		spreadFactor       sdk.Dec
	}

	tests := map[string]testcase{
		"[single asset join] even two asset pool, no spread factor": {
			tokensIn:       hundredFoo,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
		"[single asset join] uneven two asset pool, no spread factor": {
			tokensIn:       hundredFoo,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
		"[single asset join] even 3-asset pool, no spread factor": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
		"[single asset join] uneven 3-asset pool, no spread factor": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
		"[single asset join] even two asset pool, default spread factor": {
			tokensIn:       hundredFoo,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   defaultSpreadFactor,
		},
		"[single asset join] uneven two asset pool, default spread factor": {
			tokensIn:       hundredFoo,
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   defaultSpreadFactor,
		},
		"[single asset join] even 3-asset pool, default spread factor": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			spreadFactor:   defaultSpreadFactor,
		},
		"[single asset join] uneven 3-asset pool, default spread factor": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			spreadFactor:   defaultSpreadFactor,
		},
		"[single asset join] even 3-asset pool, 0.03 spread factor": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			spreadFactor:   sdk.MustNewDecFromStr("0.03"),
		},
		"[single asset join] uneven 3-asset pool, 0.03 spread factor": {
			tokensIn:       thousandAssetA,
			poolAssets:     threeUnevenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			spreadFactor:   sdk.MustNewDecFromStr("0.03"),
		},

		"[all asset join] even two asset pool, same tokenIn ratio": {
			tokensIn:       tenPercentOfTwoPoolCoins,
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
		"[all asset join] even two asset pool, different tokenIn ratio with pool": {
			tokensIn:       sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPoolRaw)), sdk.NewCoin("bar", sdk.NewInt(10+tenPercentOfTwoPoolRaw))),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
		"[all asset join] even two asset pool, different tokenIn ratio with pool, nonzero spread factor": {
			tokensIn:       sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tenPercentOfTwoPoolRaw)), sdk.NewCoin("bar", sdk.NewInt(10+tenPercentOfTwoPoolRaw))),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   defaultSpreadFactor,
		},
		"[all asset join] even two asset pool, no tokens in": {
			tokensIn:       sdk.NewCoins(),
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			spreadFactor:   sdk.ZeroDec(),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)

			// only for single asset join case
			var swapRatio sdk.Dec
			var err error
			if len(tc.tokensIn) == 1 {
				swapRatio, err = p.singleAssetJoinSpreadFactorRatio(tc.tokensIn[0].Denom)
				require.NoError(t, err)
			}

			// we join then exit the pool
			shares, _ := p.JoinPool(ctx, tc.tokensIn, tc.spreadFactor)
			tokenOut, err := p.ExitPool(ctx, shares, defaultExitFee)
			require.NoError(t, err)

			// if single asset join, we swap output tokens to input denom to test the full inverse relationship
			if len(tc.tokensIn) == 1 {
				tokenOutAmt, err := cfmm_common.SwapAllCoinsToSingleAsset(&p, ctx, tokenOut, tc.tokensIn[0].Denom, sdk.ZeroDec())
				require.NoError(t, err)
				tokenOut = sdk.NewCoins(sdk.NewCoin(tc.tokensIn[0].Denom, tokenOutAmt))
			}

			// if single asset join, we expect output token swapped into the input denom
			// to be smaller by swap ratio * 2
			var expectedTokenOut sdk.Coins
			if len(tc.tokensIn) == 1 {
				oneMinusSingleSpreadFactor := sdk.OneDec().Sub((swapRatio.Mul(tc.spreadFactor)))
				expectedAmt := (tc.tokensIn[0].Amount.ToDec().Mul(oneMinusSingleSpreadFactor)).TruncateInt()
				expectedTokenOut = sdk.NewCoins(sdk.NewCoin(tc.tokensIn[0].Denom, expectedAmt))
			} else {
				expectedTokenOut = tc.tokensIn
			}

			finalPoolLiquidity := p.GetTotalPoolLiquidity(ctx)
			require.True(t, tokenOut.IsAllLTE(expectedTokenOut), "token out %v, expected <= %v", tokenOut, expectedTokenOut)
			require.True(t, finalPoolLiquidity.IsAllGTE(tc.poolAssets))
			require.NoError(t, err)
		})
	}
}

func TestExitPool(t *testing.T) {
	tenPercentOfTwoPoolCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(int64(1000000000/10))), sdk.NewCoin("bar", sdk.NewInt(int64(1000000000/10))))
	tenPercentOfThreePoolCoins := sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(1000000/10)), sdk.NewCoin("asset/b", sdk.NewInt(1000000/10)), sdk.NewCoin("asset/c", sdk.NewInt(1000000/10)))
	tenPercentOfUnevenThreePoolCoins := sdk.NewCoins(sdk.NewCoin("asset/a", sdk.NewInt(1000000/10)), sdk.NewCoin("asset/b", sdk.NewInt(2000000/10)), sdk.NewCoin("asset/c", sdk.NewInt(3000000/10)))
	type testcase struct {
		sharesIn              sdk.Int
		initialPoolLiquidity  sdk.Coins
		scalingFactors        []uint64
		expectedPoolLiquidity sdk.Coins
		expectedTokenOut      sdk.Coins
		expectPass            bool
	}
	tests := map[string]testcase{
		"basic two-asset pool exit on even pool": {
			sharesIn:              types.InitPoolSharesSupply.Quo(sdk.NewInt(10)),
			initialPoolLiquidity:  twoEvenStablePoolAssets,
			scalingFactors:        defaultTwoAssetScalingFactors,
			expectedPoolLiquidity: twoEvenStablePoolAssets.Sub(tenPercentOfTwoPoolCoins),
			expectedTokenOut:      tenPercentOfTwoPoolCoins,
			expectPass:            true,
		},
		"basic three-asset pool exit on even pool": {
			sharesIn:              types.InitPoolSharesSupply.Quo(sdk.NewInt(10)),
			initialPoolLiquidity:  threeEvenStablePoolAssets,
			scalingFactors:        defaultThreeAssetScalingFactors,
			expectedPoolLiquidity: threeEvenStablePoolAssets.Sub(tenPercentOfThreePoolCoins),
			expectedTokenOut:      tenPercentOfThreePoolCoins,
			expectPass:            true,
		},
		"basic three-asset pool exit on uneven pool": {
			sharesIn:              types.InitPoolSharesSupply.Quo(sdk.NewInt(10)),
			initialPoolLiquidity:  threeUnevenStablePoolAssets,
			scalingFactors:        defaultThreeAssetScalingFactors,
			expectedPoolLiquidity: threeUnevenStablePoolAssets.Sub(tenPercentOfUnevenThreePoolCoins),
			expectedTokenOut:      tenPercentOfUnevenThreePoolCoins,
			expectPass:            true,
		},
		"pool exit pushes post-scaled asset below 1": {
			// attempt to exit one token when post-scaled amount is already 1 for each asset
			sharesIn:              types.InitPoolSharesSupply.Quo(sdk.NewInt(1000000)),
			initialPoolLiquidity:  threeEvenStablePoolAssets,
			scalingFactors:        []uint64{1000000 / types.ScalingFactorMultiplier, 100000 / types.ScalingFactorMultiplier, 100000 / types.ScalingFactorMultiplier},
			expectedPoolLiquidity: threeEvenStablePoolAssets,
			expectedTokenOut:      sdk.Coins{},
			expectPass:            false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.initialPoolLiquidity, tc.scalingFactors)
			tokenOut, err := p.ExitPool(ctx, tc.sharesIn, defaultExitFee)

			if tc.expectPass {
				finalPoolLiquidity := p.GetTotalPoolLiquidity(ctx)
				require.True(t, tokenOut.IsAllLTE(tc.expectedTokenOut))
				require.True(t, finalPoolLiquidity.IsAllGTE(tc.expectedPoolLiquidity))
			}
			osmoassert.ConditionalError(t, !tc.expectPass, err)
		})
	}
}

func TestValidatePoolLiquidity(t *testing.T) {
	const (
		a = "aaa"
		b = "bbb"
		c = "ccc"
		d = "ddd"
	)

	var (
		ten = sdk.NewInt(10)

		coinA = sdk.NewCoin(a, ten)
		coinB = sdk.NewCoin(b, ten)
		coinC = sdk.NewCoin(c, ten)
		coinD = sdk.NewCoin(d, ten)
	)

	tests := map[string]struct {
		liquidity      sdk.Coins
		scalingFactors []uint64
		expectError    error
	}{
		"sorted": {
			liquidity: sdk.Coins{
				coinA,
				coinB,
				coinC,
				coinD,
			},
			scalingFactors: []uint64{10, 10, 10, 10},
		},
		"unsorted - error": {
			liquidity: sdk.Coins{
				coinD,
				coinA,
				coinC,
				coinB,
			},
			scalingFactors: []uint64{10, 10, 10, 10},
			expectError: types.UnsortedPoolLiqError{ActualLiquidity: sdk.Coins{
				coinD,
				coinA,
				coinC,
				coinB,
			}},
		},
		"denoms don't match scaling factors": {
			liquidity: sdk.Coins{
				coinA,
				coinB,
				coinC,
				coinD,
			},
			scalingFactors: []uint64{10, 10},
			expectError: types.LiquidityAndScalingFactorCountMismatchError{
				LiquidityCount:     4,
				ScalingFactorCount: 2,
			},
		},
		// TODO: cover remaining edge cases by referring to the function implementation.
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validatePoolLiquidity(tc.liquidity, tc.scalingFactors)

			if tc.expectError != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSetScalingFactors(t *testing.T) {
	pk := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pk.Address())

	failPk := ed25519.GenPrivKey().PubKey()
	failAddr := sdk.AccAddress(failPk.Address())

	tests := map[string]struct {
		scalingFactors []uint64
		sender         string
		poolAssets     sdk.Coins
		expError       error
	}{
		"Sender is not scaling factor governor in pool": {
			scalingFactors: defaultTwoAssetScalingFactors,
			sender:         failAddr.String(),
			poolAssets:     twoEvenStablePoolAssets,
			expError:       types.ErrNotScalingFactorGovernor,
		},

		"Invalid scaling factor's length": {
			scalingFactors: defaultTwoAssetScalingFactors,
			sender:         addr.String(),
			poolAssets:     threeEvenStablePoolAssets,
			expError:       types.ErrInvalidScalingFactorLength,
		},
		"Invalid pool liquidity": {
			scalingFactors: []uint64{1},
			sender:         addr.String(),
			poolAssets:     sdk.NewCoins(sdk.NewInt64Coin("foo", 1000000000)),
			expError:       types.ErrTooFewPoolAssets,
		},
		"Valid set scaling for two assets in pool": {
			scalingFactors: defaultTwoAssetScalingFactors,
			sender:         addr.String(),
			poolAssets:     twoEvenStablePoolAssets,
		},
		"Valid set scaling for two uneven assets in pool": {
			scalingFactors: []uint64{2, 1},
			sender:         addr.String(),
			poolAssets:     twoUnevenStablePoolAssets,
		},
		"Valid set scaling for three assets in pool": {
			scalingFactors: defaultThreeAssetScalingFactors,
			sender:         addr.String(),
			poolAssets:     threeEvenStablePoolAssets,
		},
		"Valid set scaling for three uneven assets in pool": {
			scalingFactors: []uint64{1, 2, 3},
			sender:         addr.String(),
			poolAssets:     threeUnevenStablePoolAssets,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			pool := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)
			pool.ScalingFactorController = addr.String()
			err := pool.SetScalingFactors(ctx, tc.scalingFactors, tc.sender)
			if tc.expError != nil {
				require.Error(t, err)
				require.Equal(t, err, tc.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStableswapSpotPrice(t *testing.T) {
	type testcase struct {
		baseDenom      string
		quoteDenom     string
		poolAssets     sdk.Coins
		scalingFactors []uint64
		expectPass     bool
		expectedPrice  sdk.Dec
	}
	tests := map[string]testcase{
		"even two-asset pool": {
			baseDenom:      "foo",
			quoteDenom:     "bar",
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expectPass:     true,
		},
		"even two-asset pool with large scaling factors": {
			baseDenom:      "foo",
			quoteDenom:     "bar",
			poolAssets:     twoEvenStablePoolAssets,
			scalingFactors: []uint64{10000, 10000},
			expectPass:     true,
		},
		"even two-asset pool with different scaling factors (foo -> bar)": {
			baseDenom:      "foo",
			quoteDenom:     "bar",
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10000, 20000},
			expectedPrice:  sdk.NewDecWithPrec(5, 1),
			expectPass:     true,
		},
		"even two-asset pool with different scaling factors (bar -> foo)": {
			baseDenom:      "bar",
			quoteDenom:     "foo",
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: []uint64{10000, 20000},
			expectedPrice:  sdk.NewDec(2),
			expectPass:     true,
		},
		"uneven two-asset pool": {
			baseDenom:      "foo",
			quoteDenom:     "bar",
			poolAssets:     twoUnevenStablePoolAssets,
			scalingFactors: defaultTwoAssetScalingFactors,
			expectPass:     true,
		},
		"even three-asset pool": {
			baseDenom:      "asset/a",
			quoteDenom:     "asset/b",
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: defaultThreeAssetScalingFactors,
			expectPass:     true,
		},
		"even three-asset pool with large scaling factors": {
			baseDenom:      "asset/a",
			quoteDenom:     "asset/b",
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: []uint64{10000, 10000, 10000},
			expectPass:     true,
		},
		"even three-asset pool with different scaling factors": {
			baseDenom:      "asset/a",
			quoteDenom:     "asset/b",
			poolAssets:     threeEvenStablePoolAssets,
			scalingFactors: []uint64{500, 700, 200},
			expectPass:     true,
		},
		"uneven three-asset pool (a -> b)": {
			baseDenom:  "asset/a",
			quoteDenom: "asset/b",
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("asset/a", 10_000_000_000),
				sdk.NewInt64Coin("asset/b", 20_000_000_000),
				sdk.NewInt64Coin("asset/c", 30_000_000_000),
			),
			scalingFactors: defaultThreeAssetScalingFactors,
			expectPass:     true,
		},
		"uneven three-asset pool (b -> a)": {
			baseDenom:  "asset/b",
			quoteDenom: "asset/a",
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("asset/a", 10_000_000_000),
				sdk.NewInt64Coin("asset/b", 20_000_000_000),
				sdk.NewInt64Coin("asset/c", 30_000_000_000),
			),
			scalingFactors: defaultThreeAssetScalingFactors,
			expectPass:     true,
		},
		"uneven three-asset pool large scaling factors (a -> b)": {
			baseDenom:  "asset/a",
			quoteDenom: "asset/b",
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("asset/a", 10_000_000_000),
				sdk.NewInt64Coin("asset/b", 20_000_000_000),
				sdk.NewInt64Coin("asset/c", 30_000_000_000),
			),
			scalingFactors: []uint64{10000, 10000, 10000},
			expectPass:     true,
		},
		"uneven three-asset pool large scaling factors (b -> a)": {
			baseDenom:  "asset/b",
			quoteDenom: "asset/a",
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("asset/a", 10_000_000_000),
				sdk.NewInt64Coin("asset/b", 20_000_000_000),
				sdk.NewInt64Coin("asset/c", 30_000_000_000),
			),
			scalingFactors: []uint64{10000, 10000, 10000},
			expectPass:     true,
		},
		"uneven 3-asset pool with uneven scaling factors (a -> b)": {
			baseDenom:  "asset/a",
			quoteDenom: "asset/b",
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("asset/a", 12_345_678_910),
				sdk.NewInt64Coin("asset/b", 10_987_654_321),
				sdk.NewInt64Coin("asset/c", 8_452_398_713),
			),
			scalingFactors: []uint64{36, 578, 253},
			expectPass:     true,
		},
		"uneven 3-asset pool with uneven scaling factors (b -> a)": {
			baseDenom:  "asset/a",
			quoteDenom: "asset/b",
			poolAssets: sdk.NewCoins(
				sdk.NewInt64Coin("asset/a", 12_345_678_910),
				sdk.NewInt64Coin("asset/b", 10_987_654_321),
				sdk.NewInt64Coin("asset/c", 8_452_398_713),
			),
			scalingFactors: []uint64{36, 578, 253},
			expectPass:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.Context{}
			p := poolStructFromAssets(tc.poolAssets, tc.scalingFactors)
			spotPrice, err := p.SpotPrice(ctx, tc.quoteDenom, tc.baseDenom)

			if tc.expectPass {
				require.NoError(t, err)

				var expectedSpotPrice sdk.Dec
				if (tc.expectedPrice != sdk.Dec{}) {
					expectedSpotPrice = tc.expectedPrice
				} else {
					expectedSpotPrice, err = p.calcOutAmtGivenIn(sdk.NewInt64Coin(tc.baseDenom, 1), tc.quoteDenom, sdk.ZeroDec())
					require.NoError(t, err)
				}

				// We allow for a small geometric error due to our spot price being an approximation
				diff := (expectedSpotPrice.Sub(spotPrice)).Abs()
				errTerm := diff.Quo(sdk.MinDec(expectedSpotPrice, spotPrice))
				require.True(t, errTerm.LT(sdk.NewDecWithPrec(1, 8)), "Expected: %d, Actual: %d", expectedSpotPrice, spotPrice)

				// Pool liquidity should remain unchanged
				require.Equal(t, tc.poolAssets, p.GetTotalPoolLiquidity(ctx))
			}
			osmoassert.ConditionalError(t, !tc.expectPass, err)
		})
	}
}

func TestValidateScalingFactors(t *testing.T) {
	tests := map[string]struct {
		scalingFactors []uint64
		numAssets      int
		expectError    bool
	}{
		"number of scaling factors match number of assets": {
			numAssets:      4,
			scalingFactors: []uint64{10, 10, 10, 10},
			expectError:    false,
		},
		"number of scaling factors and assets mismatch": {
			numAssets:      3,
			scalingFactors: []uint64{10, 10, 10, 10},
			expectError:    true,
		},
		"all scaling factors equal to zero": {
			numAssets:      3,
			scalingFactors: []uint64{0, 0, 0},
			expectError:    true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateScalingFactors(tc.scalingFactors, tc.numAssets)

			if tc.expectError != false {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
