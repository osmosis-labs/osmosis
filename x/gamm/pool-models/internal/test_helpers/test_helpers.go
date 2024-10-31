package test_helpers

import (
	"math/rand"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	sdkrand "github.com/osmosis-labs/osmosis/v27/simulation/simtypes/random"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

// CfmmCommonTestSuite is the common test suite struct of Constant Function Market Maker,
// that pool-models can inherit from.
type CfmmCommonTestSuite struct {
	suite.Suite
}

func (suite *CfmmCommonTestSuite) CreateTestContext() sdk.Context {
	logger := log.NewNopLogger()
	ms := rootmulti.NewStore(dbm.NewMemDB(), logger, metrics.NewNoOpMetrics())

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger)
}

func TestCalculateAmountOutAndIn_InverseRelationship(
	t *testing.T,
	ctx sdk.Context,
	pool types.CFMMPoolI,
	assetInDenom string,
	assetOutDenom string,
	initialCalcOut int64,
	spreadFactor osmomath.Dec,
	errTolerance osmomath.ErrTolerance,
) {
	initialOut := sdk.NewInt64Coin(assetOutDenom, initialCalcOut)
	initialOutCoins := sdk.NewCoins(initialOut)

	actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, assetInDenom, spreadFactor)
	require.NoError(t, err)

	// we expect that any output less than 1 will always be rounded up
	require.True(t, actualTokenIn.Amount.GTE(osmomath.OneInt()))

	inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), assetOutDenom, spreadFactor)
	require.NoError(t, err)

	require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

	expected := initialOut.Amount.ToLegacyDec()
	actual := inverseTokenOut.Amount.ToLegacyDec()

	// If the pool is extremely imbalanced (specifically in the case of stableswap),
	// we expect there to be drastically amplified error that will fall outside our usual bounds.
	// Since these cases are effectively unusable by design, we only really care about whether
	// they are safe i.e. round correctly.
	preFeeTokenIn := actualTokenIn.Amount.ToLegacyDec().Mul((osmomath.OneDec().Sub(spreadFactor))).Ceil().TruncateInt()
	if preFeeTokenIn.Equal(osmomath.OneInt()) {
		require.True(t, actual.GT(expected))
	} else {
		if expected.Sub(actual).Abs().GT(osmomath.OneDec()) {
			compRes := errTolerance.CompareBigDec(osmomath.BigDecFromDec(expected), osmomath.BigDecFromDec(actual))
			require.True(t, compRes == 0, "expected %s, actual %s, not within error tolerance %v",
				expected, actual, errTolerance)
		}
	}
}

func TestSlippageRelationWithLiquidityIncrease(
	testname string,
	t *testing.T,
	ctx sdk.Context,
	createPoolWithLiquidity func(sdk.Context, sdk.Coins) types.CFMMPoolI,
	initLiquidity sdk.Coins,
) {
	TestSlippageRelationOutGivenIn(testname, t, ctx, createPoolWithLiquidity, initLiquidity)
	TestSlippageRelationInGivenOut(testname, t, ctx, createPoolWithLiquidity, initLiquidity)
}

func TestSlippageRelationOutGivenIn(
	testname string,
	t *testing.T,
	ctx sdk.Context,
	createPoolWithLiquidity func(sdk.Context, sdk.Coins) types.CFMMPoolI,
	initLiquidity sdk.Coins,
) {
	r := rand.New(rand.NewSource(100))
	swapInAmt := sdkrand.RandCoin(r, initLiquidity[:1])
	swapOutDenom := initLiquidity[1].Denom

	curPool := createPoolWithLiquidity(ctx, initLiquidity)
	fee := curPool.GetSpreadFactor(ctx)

	curLiquidity := initLiquidity
	curOutAmount, err := curPool.CalcOutAmtGivenIn(ctx, swapInAmt, swapOutDenom, fee)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newLiquidity := curLiquidity.Add(curLiquidity...)
		curPool = createPoolWithLiquidity(ctx, newLiquidity)

		// ensure out amount goes down as liquidity increases
		newOutAmount, err := curPool.CalcOutAmtGivenIn(ctx, swapInAmt, swapOutDenom, fee)
		require.NoError(t, err)
		require.True(t, newOutAmount.Amount.GTE(curOutAmount.Amount),
			"%s: swap with new liquidity %s yielded less than swap with old liquidity %s."+
				" Swap amount in %s. new swap out: %s, old swap out %s", testname, newLiquidity, curLiquidity,
			swapInAmt, newOutAmount, curOutAmount)

		curLiquidity, curOutAmount = newLiquidity, newOutAmount
	}
}

func TestSlippageRelationInGivenOut(
	testname string,
	t *testing.T,
	ctx sdk.Context,
	createPoolWithLiquidity func(sdk.Context, sdk.Coins) types.CFMMPoolI,
	initLiquidity sdk.Coins,
) {
	r := rand.New(rand.NewSource(100))
	swapOutAmt := sdkrand.RandCoin(r, initLiquidity[:1])
	swapInDenom := initLiquidity[1].Denom

	curPool := createPoolWithLiquidity(ctx, initLiquidity)
	fee := curPool.GetSpreadFactor(ctx)

	// we first ensure that the pool has sufficient liquidity to accommodate
	// a swap that yields `swapOutAmt` without more than doubling input reserves
	curLiquidity := initLiquidity
	for !isWithinBounds(ctx, curPool, swapOutAmt, swapInDenom, fee) {
		// increase pool liquidity by 10x
		for i, coin := range initLiquidity {
			curLiquidity[i] = sdk.NewCoin(coin.Denom, coin.Amount.Mul(osmomath.NewInt(10)))
		}
		curPool = createPoolWithLiquidity(ctx, curLiquidity)
	}

	curInAmount, err := curPool.CalcInAmtGivenOut(ctx, swapOutAmt, swapInDenom, fee)

	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newLiquidity := curLiquidity.Add(curLiquidity...)
		curPool = createPoolWithLiquidity(ctx, newLiquidity)

		// ensure required in amount goes down as liquidity increases
		newInAmount, err := curPool.CalcInAmtGivenOut(ctx, swapOutAmt, swapInDenom, fee)
		require.NoError(t, err)
		require.True(t, newInAmount.Amount.LTE(curInAmount.Amount),
			"%s: swap with new liquidity %s required greater input than swap with old liquidity %s."+
				" Swap amount out %s. new swap in: %s, old swap in %s", testname, newLiquidity, curLiquidity,
			swapOutAmt, newInAmount, curInAmount)

		curLiquidity, curInAmount = newLiquidity, newInAmount
	}
}

// returns true if the pool can accommodate an InGivenOut swap with `tokenOut` amount out, false otherwise
func isWithinBounds(ctx sdk.Context, pool types.CFMMPoolI, tokenOut sdk.Coins, tokenInDenom string, spreadFactor osmomath.Dec) (b bool) {
	b = true
	defer func() {
		if r := recover(); r != nil {
			b = false
		}
	}()
	_, err := pool.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, spreadFactor)
	if err != nil {
		b = false
	}
	return b
}

func TestCfmmCommonTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CfmmCommonTestSuite))
}
