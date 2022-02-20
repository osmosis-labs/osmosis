package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestOsmoEquivalentMultiplierSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	multipliers := suite.app.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.ctx)
	suite.Require().Len(multipliers, 0)

	// set multiplier
	suite.app.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.ctx, 1, "gamm/pool/1", sdk.NewDec(2))

	// get multiplier
	multiplier := suite.app.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// check multipliers
	expectedMultipliers := []types.OsmoEquivalentMultiplierRecord{
		{
			EpochNumber: 1,
			Denom:       "gamm/pool/1",
			Multiplier:  sdk.NewDec(2),
		},
	}
	multipliers = suite.app.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.ctx)
	suite.Require().Equal(multipliers, expectedMultipliers)

	params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochstypes.EpochInfo{
		Identifier:   params.RefreshEpochIdentifier,
		CurrentEpoch: 2,
	})

	// test last epoch price
	multiplier = suite.app.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// delete multiplier
	suite.app.SuperfluidKeeper.DeleteOsmoEquivalentMultiplier(suite.ctx, "gamm/pool/1")

	// get multiplier
	multiplier = suite.app.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(0))

	// check multipliers
	multipliers = suite.app.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.ctx)
	suite.Require().Len(multiplier, 0)

	// test last epoch price
	multiplier = suite.app.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(0))
}
