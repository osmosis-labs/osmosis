package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestOsmoEquivalentMultiplierSetGetDeleteFlow() {
	s.SetupTest()

	// initial check
	multipliers := s.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(s.Ctx)
	s.Require().Len(multipliers, 0)

	// set multiplier
	s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, 1, DefaultGammAsset, osmomath.NewDec(2))

	// get multiplier
	multiplier := s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, DefaultGammAsset)
	s.Require().Equal(multiplier, osmomath.NewDec(2))

	// check multipliers
	expectedMultipliers := []types.OsmoEquivalentMultiplierRecord{
		{
			EpochNumber: 1,
			Denom:       DefaultGammAsset,
			Multiplier:  osmomath.NewDec(2),
		},
	}
	multipliers = s.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(s.Ctx)
	s.Require().Equal(multipliers, expectedMultipliers)

	// test last epoch price
	multiplier = s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, DefaultGammAsset)
	s.Require().Equal(multiplier, osmomath.NewDec(2))

	// delete multiplier
	s.App.SuperfluidKeeper.DeleteOsmoEquivalentMultiplier(s.Ctx, DefaultGammAsset)

	// get multiplier
	multiplier = s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, DefaultGammAsset)
	s.Require().Equal(multiplier, osmomath.NewDec(0))

	// check multipliers
	multipliers = s.App.SuperfluidKeeper.GetAllOsmoEquivalentMultipliers(s.Ctx)
	s.Require().Len(multipliers, 0)

	// test last epoch price
	multiplier = s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, DefaultGammAsset)
	s.Require().Equal(multiplier, osmomath.NewDec(0))
}

func (s *KeeperTestSuite) TestGetSuperfluidOSMOTokens() {
	s.SetupTest()
	minRiskFactor := s.App.SuperfluidKeeper.GetParams(s.Ctx).MinimumRiskFactor
	poolCoins := sdk.NewCoins(sdk.NewCoin("stake", osmomath.NewInt(1000000000000000000)), sdk.NewCoin("foo", osmomath.NewInt(1000000000000000000)))
	s.PrepareBalancerPoolWithCoins(poolCoins...)
	s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("stake", "foo")

	gammShareDenom := DefaultGammAsset
	clShareDenom := cltypes.GetConcentratedLockupDenomFromPoolId(2)

	multiplier := osmomath.NewDec(2)
	testAmount := osmomath.NewInt(100)
	epoch := int64(1)

	// Set multiplier
	s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, epoch, gammShareDenom, multiplier)

	// Get multiplier
	multiplier = s.App.SuperfluidKeeper.GetOsmoEquivalentMultiplier(s.Ctx, gammShareDenom)
	s.Require().Equal(multiplier, osmomath.NewDec(2))

	// Should get error since asset is not superfluid enabled
	osmoTokens, err := s.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(s.Ctx, gammShareDenom, testAmount)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrNonSuperfluidAsset)
	s.Require().Equal(osmoTokens, osmomath.NewInt(0))

	// Set gamm share as superfluid
	superfluidGammAsset := types.SuperfluidAsset{
		Denom:     gammShareDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	}
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, superfluidGammAsset)
	s.Require().NoError(err)

	// Reset multiplier
	s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, epoch, gammShareDenom, multiplier)

	// Get superfluid OSMO tokens
	osmoTokens, err = s.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(s.Ctx, gammShareDenom, testAmount)
	s.Require().NoError(err)

	// Adjust result with risk factor
	osmoTokensRiskAdjusted := s.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(s.Ctx, osmoTokens)

	// Check result
	s.Require().Equal(testAmount.ToLegacyDec().Mul(minRiskFactor).TruncateInt().String(), osmoTokensRiskAdjusted.String())

	// Set cl share as superfluid
	superfluidClAsset := types.SuperfluidAsset{
		Denom:     clShareDenom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	}
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, superfluidClAsset)
	s.Require().NoError(err)

	// Reset multiplier
	s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, epoch, clShareDenom, multiplier)

	// Get superfluid OSMO tokens
	osmoTokens, err = s.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(s.Ctx, clShareDenom, testAmount)
	s.Require().NoError(err)

	// Adjust result with risk factor
	osmoTokensRiskAdjusted = s.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(s.Ctx, osmoTokens)

	// Check result
	s.Require().Equal(testAmount.ToLegacyDec().Mul(minRiskFactor).TruncateInt().String(), osmoTokensRiskAdjusted.String())
}
