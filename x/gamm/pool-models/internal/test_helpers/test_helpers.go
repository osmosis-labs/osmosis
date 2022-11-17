package test_helpers

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	sdkrand "github.com/osmosis-labs/osmosis/v12/simulation/simtypes/random"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// CfmmCommonTestSuite is the common test suite struct of Constant Function Market Maker,
// that pool-models can inherit from.
type CfmmCommonTestSuite struct {
	suite.Suite
}

func (suite *CfmmCommonTestSuite) CreateTestContext() sdk.Context {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger)

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger)
}

func TestCalculateAmountOutAndIn_InverseRelationship(
	t *testing.T,
	ctx sdk.Context,
	pool types.PoolI,
	assetInDenom string,
	assetOutDenom string,
	initialCalcOut int64,
	swapFee sdk.Dec,
) {
	initialOut := sdk.NewInt64Coin(assetOutDenom, initialCalcOut)
	initialOutCoins := sdk.NewCoins(initialOut)

	actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, assetInDenom, swapFee)
	require.NoError(t, err)

	// we expect that any output less than 1 will always be rounded up
	require.True(t, actualTokenIn.Amount.GTE(sdk.OneInt()))

	inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), assetOutDenom, swapFee)
	require.NoError(t, err)

	require.Equal(t, initialOut.Denom, inverseTokenOut.Denom)

	expected := initialOut.Amount.ToDec()
	actual := inverseTokenOut.Amount.ToDec()

	// If the pool is extremely imbalanced (specifically in the case of stableswap),
	// we expect there to be drastically amplified error that will fall outside our usual bounds.
	// Since these cases are effectively unusable by design, we only really care about whether
	// they are safe i.e. round correctly.
	preFeeTokenIn := actualTokenIn.Amount.ToDec().Mul((sdk.OneDec().Sub(swapFee))).Ceil().TruncateInt()
	if preFeeTokenIn.Equal(sdk.OneInt()) {
		require.True(t, actual.GT(expected))
	} else {
		// allow a rounding error of up to 1 for this relation
		// TODO: Ensure rounding is correct
		tol := sdk.NewDec(1)
		osmoassert.DecApproxEq(t, expected, actual, tol)
	}
}

func TestSlippageRelationWithLiquidityIncrease(
	testname string,
	t *testing.T,
	ctx sdk.Context,
	createPoolWithLiquidity func(sdk.Context, sdk.Coins) types.PoolI,
	initLiquidity sdk.Coins) {
	r := rand.New(rand.NewSource(100))
	swapInAmt := sdkrand.RandSubsetCoins(r, initLiquidity[:1])
	swapOutDenom := initLiquidity[1].Denom

	curPool := createPoolWithLiquidity(ctx, initLiquidity)
	fee := curPool.GetSwapFee(ctx)

	curLiquidity := initLiquidity
	curOutAmount, err := curPool.CalcOutAmtGivenIn(ctx, swapInAmt, swapOutDenom, fee)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
		newLiquidity := curLiquidity.Add(curLiquidity...)
		curPool = createPoolWithLiquidity(ctx, newLiquidity)
		newOutAmount, err := curPool.CalcOutAmtGivenIn(ctx, swapInAmt, swapOutDenom, fee)
		require.NoError(t, err)
		require.True(t, newOutAmount.Amount.GTE(curOutAmount.Amount),
			"%s: swap with new liquidity %s yielded less than swap with old liquidity %s."+
				" Swap amount in %s. new swap out: %s, old swap out %s", testname, newLiquidity, curLiquidity,
			swapInAmt, newOutAmount, curOutAmount)

		curLiquidity, curOutAmount = newLiquidity, newOutAmount
	}
}

func TestCfmmCommonTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CfmmCommonTestSuite))
}
