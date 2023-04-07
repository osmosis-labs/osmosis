package v16_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	v16 "github.com/osmosis-labs/osmosis/v15/app/upgrades/v16"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
)

type UpgradeTestSuite struct {
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

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestGetGaugesForCFMMPool() {
	tests := map[string]struct {
		poolId         uint64
		expectedGauges incentivestypes.Gauge
		expectError    error
	}{
		"valid pool id - gauges created": {
			poolId: validPoolId,
		},
		"invalid pool id - error": {
			poolId:      validPoolId + 1,
			expectError: poolincentivestypes.NoGaugeAssociatedWithPoolError{PoolId: 2, Duration: time.Hour},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			suite.PrepareBalancerPool()

			gauges, err := v16.GetGaugesForCFMMPool(suite.Ctx, *suite.App.IncentivesKeeper, *suite.App.PoolIncentivesKeeper, tc.poolId)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}

			suite.Require().NoError(err)

			// Create 3 gauges for each lockable duration.
			suite.Require().Equal(3, len(gauges))
			for i, lockableDuration := range suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx) {
				suite.Require().Equal(uint64(i+1), gauges[i].Id)
				suite.Require().Equal(lockableDuration, gauges[i].DistributeTo.Duration)
				suite.Require().True(gauges[i].IsActiveGauge(suite.Ctx.BlockTime()))
			}
		})
	}
}

func (suite *UpgradeTestSuite) TestCreateConcentratedPoolFromCFMM() {
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

func (suite *UpgradeTestSuite) TestCreateCanonicalConcentratedLiuqidityPoolAndMigrationLink() {
	suite.Setup()

	distributionEpochDuration := suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx).Duration

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
			expectError:           v16.CouldNotFindGaugeToRedirectError{DistributionEpochDuration: distributionEpochDuration},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			validDurations := []time.Duration{
				distributionEpochDuration,
				time.Hour, // random value
			}

			if !tc.setupInvalidDuraitons {
				// Overwrite default lockable durations that do not have the distribution epoch duration
				suite.App.PoolIncentivesKeeper.SetLockableDurations(suite.Ctx, validDurations)
				suite.App.IncentivesKeeper.SetLockableDurations(suite.Ctx, validDurations)
			}

			balancerId := suite.PrepareBalancerPoolWithCoins(tc.poolLiquidity...)

			balancerPool, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, balancerId)
			suite.Require().NoError(err)

			// Get balance gauges.
			gaugeToRedirect, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerPool.GetId(), distributionEpochDuration)
			suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, poolincentivestypes.DistrInfo{
				TotalWeight: sdk.NewInt(100),
				Records: []poolincentivestypes.DistrRecord{
					{
						GaugeId: gaugeToRedirect,
						Weight:  sdk.NewInt(100),
					},
				},
			})

			err = v16.CreateCanonicalConcentratedLiuqidityPoolAndMigrationLink(suite.Ctx, tc.cfmmPoolIdToLinkWith, tc.desiredDenom0, &suite.App.AppKeepers)

			if tc.expectError != nil {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)

			// Get the new concentrated pool.
			clPoolInState, err := suite.App.PoolManagerKeeper.GetPool(suite.Ctx, validPoolId+1)
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
		})
	}
}
