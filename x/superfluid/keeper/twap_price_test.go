package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestEpochOsmoEquivalentTWAPSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	twaps := suite.app.SuperfluidKeeper.GetAllEpochOsmoEquivalentTWAPs(suite.ctx, 1)
	suite.Require().Len(twaps, 0)

	// set twap
	suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 1, "gamm/pool/1", sdk.NewDec(2))

	// get twap
	twap := suite.app.SuperfluidKeeper.GetEpochOsmoEquivalentTWAP(suite.ctx, 1, "gamm/pool/1")
	suite.Require().Equal(twap, sdk.NewDec(2))

	// check twaps
	expectedTwaps := []types.EpochOsmoEquivalentTWAP{
		{
			EpochNumber:    1,
			Denom:          "gamm/pool/1",
			EpochTwapPrice: sdk.NewDec(2),
		},
	}
	twaps = suite.app.SuperfluidKeeper.GetAllEpochOsmoEquivalentTWAPs(suite.ctx, 1)
	suite.Require().Equal(twaps, expectedTwaps)
	twaps = suite.app.SuperfluidKeeper.GetAllOsmoEquivalentTWAPs(suite.ctx)
	suite.Require().Equal(twaps, expectedTwaps)

	params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   params.RefreshEpochIdentifier,
		CurrentEpoch: 2,
	})

	// test last epoch price
	twapT := suite.app.SuperfluidKeeper.GetLastEpochOsmoEquivalentTWAP(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(twapT.Denom, "gamm/pool/1")
	suite.Require().Equal(twapT.EpochNumber, int64(1))
	suite.Require().Equal(twapT.EpochTwapPrice, sdk.NewDec(2))

	// delete twap
	suite.app.SuperfluidKeeper.DeleteEpochOsmoEquivalentTWAP(suite.ctx, 1, "gamm/pool/1")

	// get twap
	twap = suite.app.SuperfluidKeeper.GetEpochOsmoEquivalentTWAP(suite.ctx, 1, "gamm/pool/1")
	suite.Require().Equal(twap, sdk.NewDec(0))

	// check twaps
	twaps = suite.app.SuperfluidKeeper.GetAllEpochOsmoEquivalentTWAPs(suite.ctx, 1)
	suite.Require().Len(twaps, 0)
	twaps = suite.app.SuperfluidKeeper.GetAllOsmoEquivalentTWAPs(suite.ctx)
	suite.Require().Len(twaps, 0)

	// test last epoch price
	twapT = suite.app.SuperfluidKeeper.GetLastEpochOsmoEquivalentTWAP(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(twapT.Denom, "gamm/pool/1")
	suite.Require().Equal(twapT.EpochNumber, int64(1))
	suite.Require().Equal(twapT.EpochTwapPrice, sdk.NewDec(0))

	// test current epoch price
	twapT = suite.app.SuperfluidKeeper.GetCurrentEpochOsmoEquivalentTWAP(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(twapT.Denom, "gamm/pool/1")
	suite.Require().Equal(twapT.EpochNumber, int64(2))
	suite.Require().Equal(twapT.EpochTwapPrice, sdk.NewDec(0))

	suite.app.SuperfluidKeeper.SetEpochOsmoEquivalentTWAP(suite.ctx, 2, "gamm/pool/1", sdk.NewDec(20))
	twapT = suite.app.SuperfluidKeeper.GetCurrentEpochOsmoEquivalentTWAP(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(twapT.Denom, "gamm/pool/1")
	suite.Require().Equal(twapT.EpochNumber, int64(2))
	suite.Require().Equal(twapT.EpochTwapPrice, sdk.NewDec(20))
}
