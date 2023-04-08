package v16_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	v16 "github.com/osmosis-labs/osmosis/v15/app/upgrades/v16"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
)

type ConcentratedUpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

const (
	validPoolId = uint64(1)
)

var (
	defaultAmount = sdk.NewInt(100)
	uosmoDenom    = "uosmo"
	uusdDenom     = "uusd"
	coinA         = sdk.NewCoin(uosmoDenom, defaultAmount)
	coinB         = sdk.NewCoin("uatom", defaultAmount)
	coinC         = sdk.NewCoin(uusdDenom, defaultAmount)
)

func (suite *ConcentratedUpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestConcentratedUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(ConcentratedUpgradeTestSuite))
}

func (suite *ConcentratedUpgradeTestSuite) TestCreateConcentratedPoolFromCFMM() {
	tests := map[string]struct {
		poolLiquidity sdk.Coins

		cfmmPoolIdToLinkWith uint64
		desiredDenom0        string
		expectError          error
	}{
		"success": {
			poolLiquidity:        sdk.NewCoins(coinA, coinB),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        uosmoDenom,
		},
		"error: invalid denom 0": {
			poolLiquidity:        sdk.NewCoins(coinA, coinB),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        uusdDenom,
			expectError:          v16.NoDesiredDenomInPoolError{uusdDenom},
		},
		"error: pool with 3 assets, must have two": {
			poolLiquidity:        sdk.NewCoins(coinA, coinB, coinC),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        uusdDenom,
			expectError:          v16.ErrMustHaveTwoDenoms,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			balancerId := suite.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			balancerPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, balancerId)
			suite.Require().NoError(err)

			clPoolReturned, err := v16.CreateConcentratedPoolFromCFMM(suite.Ctx, tc.cfmmPoolIdToLinkWith, tc.desiredDenom0, *suite.App.AccountKeeper, *suite.App.GAMMKeeper, *suite.App.PoolManagerKeeper)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().Nil(clPoolReturned)
				return
			}
			suite.Require().NoError(err)

			// Validate that pool saved in state is the same as the one returned
			clPoolInState, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, clPoolReturned.GetId())
			suite.Require().NoError(err)
			suite.Require().Equal(clPoolReturned, clPoolInState)

			// Validate that CL and balancer pools have the same denoms
			balancerDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, balancerPool.GetId())
			suite.Require().NoError(err)

			clDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, clPoolReturned.GetId())
			suite.Require().NoError(err)

			suite.Require().Equal(balancerDenoms, clDenoms)
		})
	}
}

func (suite *ConcentratedUpgradeTestSuite) TestCreateCanonicalConcentratedLiuqidityPoolAndMigrationLink() {
	suite.Setup()

	locableDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx)
	longestLockableDuration := locableDurations[len(locableDurations)-1]

	tests := map[string]struct {
		poolLiquidity sdk.Coins

		cfmmPoolIdToLinkWith  uint64
		desiredDenom0         string
		setupInvalidDuraitons bool
		expectError           error
	}{
		"success": {
			poolLiquidity:        sdk.NewCoins(coinA, coinB),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        uosmoDenom,
		},
		"error: invalid denom 0": {
			poolLiquidity:        sdk.NewCoins(coinA, coinB),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        uusdDenom,
			expectError:          v16.NoDesiredDenomInPoolError{uusdDenom},
		},
		"error: pool with 3 assets, must have two": {
			poolLiquidity:        sdk.NewCoins(coinA, coinB, coinC),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        uusdDenom,
			expectError:          v16.ErrMustHaveTwoDenoms,
		},
		"error: invalid denom durations": {
			poolLiquidity:         sdk.NewCoins(coinA, coinB),
			cfmmPoolIdToLinkWith:  validPoolId,
			desiredDenom0:         uosmoDenom,
			setupInvalidDuraitons: true,
			expectError:           v16.ErrNoGaugeToRedirect,
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			if tc.setupInvalidDuraitons {
				// Overwrite default lockable durations.
				suite.App.PoolIncentivesKeeper.SetLockableDurations(suite.Ctx, []time.Duration{})
			}

			balancerId := suite.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			// Another pool for testing that its gauge linkes are unchanged
			balancerId2 := suite.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			balancerPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, balancerId)
			suite.Require().NoError(err)

			// Get balance gauges.
			gaugeToRedirect, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerPool.GetId(), longestLockableDuration)

			gaugeToNotRedeirect, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerId2, longestLockableDuration)

			originalDistrInfo := poolincentivestypes.DistrInfo{
				TotalWeight: sdk.NewInt(100),
				Records: []poolincentivestypes.DistrRecord{
					{
						GaugeId: gaugeToRedirect,
						Weight:  sdk.NewInt(50),
					},
					{
						GaugeId: gaugeToNotRedeirect,
						Weight:  sdk.NewInt(50),
					},
				},
			}
			suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, originalDistrInfo)

			err = v16.CreateCanonicalConcentratedLiuqidityPoolAndMigrationLink(suite.Ctx, tc.cfmmPoolIdToLinkWith, tc.desiredDenom0, &suite.App.AppKeepers)

			if tc.expectError != nil {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)

			// Get the new concentrated pool.
			// Note, + 2 becuse we create 2 balancer pools during test setup, and 1 concentrated pool during migration.
			clPoolInState, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, validPoolId+2)
			suite.Require().NoError(err)

			// Validate that CL and balancer pools have the same denoms
			balancerDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, balancerPool.GetId())
			suite.Require().NoError(err)

			clDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, clPoolInState.GetId())
			suite.Require().NoError(err)

			suite.Require().Equal(balancerDenoms, clDenoms)

			// Validate that CFMM gauge is linked to the new concentrated pool.
			concentratedPoolGaugeId, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, clPoolInState.GetId(), suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx).Duration)
			suite.Require().NoError(err)

			distrInfo := suite.App.PoolIncentivesKeeper.GetDistrInfo(suite.Ctx)
			suite.Require().Equal(distrInfo.Records[0].GaugeId, concentratedPoolGaugeId)

			// Validate that distribution record from another pool is not redirected.
			suite.Require().Equal(distrInfo.Records[1].GaugeId, gaugeToNotRedeirect)

			// Validate migration record.
			migrationInfo := suite.App.GAMMKeeper.GetMigrationInfo(suite.Ctx)
			suite.Require().Equal(migrationInfo, gammtypes.MigrationRecords{
				BalancerToConcentratedPoolLinks: []gammtypes.BalancerToConcentratedPoolLink{
					{
						BalancerPoolId: balancerId,
						ClPoolId:       clPoolInState.GetId(),
					},
				},
			})

			// Validate that old gauge still exist
			_, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeToRedirect)
			suite.Require().NoError(err)
		})
	}
}
