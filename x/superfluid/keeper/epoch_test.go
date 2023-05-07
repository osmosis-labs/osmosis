package keeper_test

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestUpdateOsmoEquivalentMultipliers() {
	testCases := []struct {
		name               string
		asset              types.SuperfluidAsset
		expectedMultiplier sdk.Dec
		removeStakingAsset bool
		poolDoesNotExist   bool
		expectedError      error
	}{
		{
			name:               "update LP token Osmo equivalent successfully",
			asset:              types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedMultiplier: sdk.MustNewDecFromStr("0.01"),
		},
		{
			name:             "update LP token Osmo equivalent with pool unexpectedly deleted",
			asset:            types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			poolDoesNotExist: true,
			expectedError:    gammtypes.PoolDoesNotExistError{PoolId: 1},
		},
		{
			name:               "update LP token Osmo equivalent with pool unexpectedly removed Osmo",
			asset:              types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			removeStakingAsset: true,
			expectedError:      errors.New("pool 1 has zero OSMO amount"),
		},
		{
			name:               "update concentrated share Osmo equivalent successfully",
			asset:              types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			expectedMultiplier: sdk.MustNewDecFromStr("1"),
		},
		{
			name:             "update concentrated share Osmo equivalent with pool unexpectedly deleted",
			asset:            types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			poolDoesNotExist: true,
			expectedError:    cltypes.PoolNotFoundError{PoolId: 1},
		},
		{
			name:               "update concentrated share Osmo equivalent with pool unexpectedly removed Osmo",
			asset:              types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			removeStakingAsset: true,
			expectedError:      errors.New("pool has unexpectedly removed OSMO as one of its underlying assets"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := suite.Ctx
			superfluidKeeper := suite.App.SuperfluidKeeper

			// Switch the default staking denom to something else if the test case requires it
			stakeDenom := suite.App.StakingKeeper.BondDenom(ctx)
			if tc.removeStakingAsset {
				stakeDenom = "bar"
			}
			poolCoins := sdk.NewCoins(sdk.NewCoin(stakeDenom, sdk.NewInt(1000000000000000000)), sdk.NewCoin("foo", sdk.NewInt(1000000000000000000)))

			// Ensure that the multiplier is zero before the test
			multiplier := superfluidKeeper.GetOsmoEquivalentMultiplier(ctx, tc.asset.Denom)
			suite.Require().Equal(multiplier, sdk.ZeroDec())

			// Create the respective pool if the test case requires it
			if !tc.poolDoesNotExist {
				if tc.asset.AssetType == types.SuperfluidAssetTypeLPShare {
					suite.PrepareBalancerPoolWithCoins(poolCoins...)
				} else if tc.asset.AssetType == types.SuperfluidAssetTypeConcentratedShare {
					suite.PrepareConcentratedPoolWithCoinsAndLockedFullRangePosition(stakeDenom, "foo")
				}
			}

			// System under test
			err := superfluidKeeper.UpdateOsmoEquivalentMultipliers(ctx, tc.asset, 1)

			if tc.expectedError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorContains(err, tc.expectedError.Error())

				// Ensure unwind superfluid asset is called
				// Check that multiplier was not set
				multiplier := superfluidKeeper.GetOsmoEquivalentMultiplier(ctx, tc.asset.Denom)
				suite.Require().Equal(multiplier, sdk.ZeroDec())
				// Check that the asset was deleted
				_, err := superfluidKeeper.GetSuperfluidAsset(ctx, tc.asset.Denom)
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				// Check that multiplier was set correctly
				multiplier := superfluidKeeper.GetOsmoEquivalentMultiplier(ctx, tc.asset.Denom)
				suite.Require().NotEqual(multiplier, sdk.ZeroDec())
			}
		})
	}
}
