package twap_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

// TestCreatePoolFlow tests that upon a pool being created,
// we have made the correct store entries.
func (suite *TestSuite) TestCreateTwoAssetPoolFlow() {
	poolLiquidity := sdk.NewCoins(sdk.NewInt64Coin("token/A", 100), sdk.NewInt64Coin("token/B", 100))
	poolId := suite.PrepareUni2PoolWithAssets(poolLiquidity[0], poolLiquidity[1])

	expectedTwap := types.NewTwapRecord(suite.App.GAMMKeeper, suite.Ctx, poolId, "token/B", "token/A")

	twap, err := suite.twapkeeper.GetMostRecentTWAP(suite.Ctx, poolId, "token/B", "token/A")
	suite.Require().NoError(err)
	suite.Require().Equal(expectedTwap, twap)

	twap, err = suite.twapkeeper.GetRecordAtOrBeforeTime(suite.Ctx, poolId, suite.Ctx.BlockTime(), "token/B", "token/A")
	suite.Require().NoError(err)
	suite.Require().Equal(expectedTwap, twap)
}
