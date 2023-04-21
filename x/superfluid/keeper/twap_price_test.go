package keeper_test

import (
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestOsmoEquivalentMultiplierSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	multipliers := suite.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.Ctx)
	suite.Require().Len(multipliers, 0)

	// set multiplier
	suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, 1, DefaultGammAsset, sdk.NewDec(2))

	// get multiplier
	multiplier := suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, DefaultGammAsset)
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// check multipliers
	expectedMultipliers := []types.OsmoEquivalentMultiplierRecord{
		{
			EpochNumber: 1,
			Denom:       DefaultGammAsset,
			Multiplier:  sdk.NewDec(2),
		},
	}
	multipliers = suite.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.Ctx)
	suite.Require().Equal(multipliers, expectedMultipliers)

	// test last epoch price
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, DefaultGammAsset)
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// delete multiplier
	suite.App.SuperfluidKeeper.DeleteOsmoEquivalentMultiplier(suite.Ctx, DefaultGammAsset)

	// get multiplier
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, DefaultGammAsset)
	suite.Require().Equal(multiplier, sdk.NewDec(0))

	// check multipliers
	multipliers = suite.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(suite.Ctx)
	suite.Require().Len(multipliers, 0)

	// test last epoch price
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, DefaultGammAsset)
	suite.Require().Equal(multiplier, sdk.NewDec(0))
}

func (suite *KeeperTestSuite) TestGetSuperfluidOSMOTokens() {
	suite.SetupTest()
	minRiskFactor := suite.App.SuperfluidKeeper.GetParams(suite.Ctx).MinimumRiskFactor
	poolCoins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000000000000000000)), sdk.NewCoin("foo", sdk.NewInt(1000000000000000000)))
	suite.PrepareBalancerPoolWithCoins(poolCoins...)
	suite.PrepareConcentratedPoolWithCoinsAndFullRangePosition("stake", "foo")

	gammShareDenom := DefaultGammAsset
	clShareDenom := cltypes.GetConcentratedLockupDenomFromPoolId(2)

	multiplier := sdk.NewDec(2)
	testAmount := sdk.NewInt(100)
	epoch := int64(1)

	// Set multiplier
	suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, epoch, gammShareDenom, multiplier)

	// Get multiplier
	multiplier = suite.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(suite.Ctx, gammShareDenom)
	suite.Require().Equal(multiplier, sdk.NewDec(2))

	// Should get error since asset is not superfluid enabled
	osmoTokens, err := suite.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(suite.Ctx, gammShareDenom, testAmount)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrNonSuperfluidAsset)
	suite.Require().Equal(osmoTokens, sdk.NewInt(0))

	// Set gamm share as superfluid
	superfluidGammAsset := types.SuperfluidAsset{
		Denom:     gammShareDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	err = suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, superfluidGammAsset)
	suite.Require().NoError(err)

	// Reset multiplier
	suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, epoch, gammShareDenom, multiplier)

	// Get superfluid OSMO tokens
	osmoTokens, err = suite.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(suite.Ctx, gammShareDenom, testAmount)
	suite.Require().NoError(err)

	// Adjust result with risk factor
	osmoTokensRiskAdjusted := suite.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.Ctx, osmoTokens)

	// Check result
	suite.Require().Equal(testAmount.ToDec().Mul(minRiskFactor).TruncateInt().String(), osmoTokensRiskAdjusted.String())

	// Set cl share as superfluid
	superfluidClAsset := types.SuperfluidAsset{
		Denom:     clShareDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	}
	err = suite.App.SuperfluidKeeper.AddNewSuperfluidAsset(suite.Ctx, superfluidClAsset)
	suite.Require().NoError(err)

	// Reset multiplier
	suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, epoch, clShareDenom, multiplier)

	// Get superfluid OSMO tokens
	osmoTokens, err = suite.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(suite.Ctx, clShareDenom, testAmount)
	suite.Require().NoError(err)

	// Adjust result with risk factor
	osmoTokensRiskAdjusted = suite.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.Ctx, osmoTokens)

	// Check result
	suite.Require().Equal(testAmount.ToDec().Mul(minRiskFactor).TruncateInt().String(), osmoTokensRiskAdjusted.String())
}
