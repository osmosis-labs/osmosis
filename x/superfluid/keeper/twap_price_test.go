package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v8/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestOsmoEquivalentMultiplierSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	multipliers := suite.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.Ctx)
	suite.Require().Len(multipliers, 0)

	// set multiplier
	suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, 1, "gamm/pool/1", sdk.NewDec(2))

	// get multiplier
	multiplier := suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// check multipliers
	expectedMultipliers := []types.OsmoEquivalentMultiplierRecord{
		{
			EpochNumber: 1,
			Denom:       "gamm/pool/1",
			Multiplier:  sdk.NewDec(2),
		},
	}
	multipliers = suite.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.Ctx)
	suite.Require().Equal(multipliers, expectedMultipliers)

	epochIdentifier := suite.App.SuperfluidKeeper.GetEpochIdentifier(suite.Ctx)
	suite.App.EpochsKeeper.SetEpochInfo(suite.Ctx, epochstypes.EpochInfo{
		Identifier:   epochIdentifier,
		CurrentEpoch: 2,
	})

	// test last epoch price
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// delete multiplier
	suite.App.SuperfluidKeeper.DeleteOsmoEquivalentMultiplier(suite.Ctx, "gamm/pool/1")

	// get multiplier
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(0))

	// check multipliers
	multipliers = suite.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.Ctx)
	suite.Require().Len(multipliers, 0)

	// test last epoch price
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, "gamm/pool/1")
	suite.Require().Equal(multiplier, sdk.NewDec(0))
}
