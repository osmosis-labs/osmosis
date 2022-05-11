package test_helpers

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

type CfmmCommonTestSuite struct {
	suite.Suite
}

func (suite *CfmmCommonTestSuite) TestCalculateAmountOutAndIn_InverseRelationship(
	ctx sdk.Context,
	pool types.PoolI,
	assetInDenom string,
	assetOutDenom string,
	initialCalcOut int64,
	swapFee sdk.Dec) {
	initialOut := sdk.NewInt64Coin(assetOutDenom, initialCalcOut)
	initialOutCoins := sdk.NewCoins(initialOut)

	actualTokenIn, err := pool.CalcInAmtGivenOut(ctx, initialOutCoins, assetInDenom, swapFee)
	suite.Require().NoError(err)

	inverseTokenOut, err := pool.CalcOutAmtGivenIn(ctx, sdk.NewCoins(actualTokenIn), assetOutDenom, swapFee)
	suite.Require().NoError(err)

	suite.Require().Equal(initialOut.Denom, inverseTokenOut.Denom)

	expected := initialOut.Amount.ToDec()
	actual := inverseTokenOut.Amount.ToDec()

	// allow a rounding error of up to 1 for this relation
	tol := sdk.NewDec(1)
	_, approxEqual, _, _, _ := osmoutils.DecApproxEq(suite.T(), expected, actual, tol)
	suite.Require().True(approxEqual)
}

func TestCfmmCommonTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CfmmCommonTestSuite))
}
