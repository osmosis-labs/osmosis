package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochOsmoEquivalentTWAPSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	twaps := suite.app.SuperfluidKeeper.GetAllEpochOsmoEquivalentTWAPs(suite.ctx, 1)
	suite.Require().Len(twaps, 0)

	// set twap
	suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 1, "lptoken", sdk.NewDec(2))

	// get twap
	twap := suite.app.SuperfluidKeeper.GetEpochOsmoEquivalentTWAP(suite.ctx, 1, "lptoken")
	suite.Require().Equal(twap, sdk.NewDec(2))

	// check twaps
	twaps = suite.app.SuperfluidKeeper.GetAllEpochOsmoEquivalentTWAPs(suite.ctx, 1)
	suite.Require().Len(twaps, 1)
	twaps = suite.app.SuperfluidKeeper.GetAllOsmoEquivalentTWAPs(suite.ctx)
	suite.Require().Len(twaps, 1)

	params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   params.RefreshEpochIdentifier,
		CurrentEpoch: 2,
	})

	// test last epoch price
	twapT := suite.app.SuperfluidKeeper.GetLastEpochOsmoEquivalentTWAP(suite.ctx, "lptoken")
	suite.Require().Equal(twapT.Denom, "lptoken")
	suite.Require().Equal(twapT.Epoch, int64(1))
	suite.Require().Equal(twapT.EpochTwapPrice, sdk.NewDec(2))

	// delete twap
	suite.app.SuperfluidKeeper.DeleteEpochOsmoEquivalentTWAP(suite.ctx, 1, "lptoken")

	// get twap
	twap = suite.app.SuperfluidKeeper.GetEpochOsmoEquivalentTWAP(suite.ctx, 1, "lptoken")
	suite.Require().Equal(twap, sdk.NewDec(0))

	// check twaps
	twaps = suite.app.SuperfluidKeeper.GetAllEpochOsmoEquivalentTWAPs(suite.ctx, 1)
	suite.Require().Len(twaps, 0)
	twaps = suite.app.SuperfluidKeeper.GetAllOsmoEquivalentTWAPs(suite.ctx)
	suite.Require().Len(twaps, 0)

	// test last epoch price
	twapT = suite.app.SuperfluidKeeper.GetLastEpochOsmoEquivalentTWAP(suite.ctx, "lptoken")
	suite.Require().Equal(twapT.Denom, "lptoken")
	suite.Require().Equal(twapT.Epoch, int64(1))
	suite.Require().Equal(twapT.EpochTwapPrice, sdk.NewDec(0))
}
