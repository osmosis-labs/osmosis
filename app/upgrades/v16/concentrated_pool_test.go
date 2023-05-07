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
	defaultAmount     = sdk.NewInt(100)
	desiredDenom0     = v16.DesiredDenom0
	desiredDenom0Coin = sdk.NewCoin(desiredDenom0, defaultAmount)
	daiCoin           = sdk.NewCoin(v16.DAIIBCDenom, defaultAmount)
	usdcCoin          = sdk.NewCoin(v16.USDCIBCDenom, defaultAmount)
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
		expectedDenoms       []string
		expectError          error
	}{
		"success": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        desiredDenom0,
			expectedDenoms:       []string{desiredDenom0, daiCoin.Denom},
		},
		"error: invalid denom 0": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        v16.USDCIBCDenom,
			expectError:          v16.NoDesiredDenomInPoolError{v16.USDCIBCDenom},
		},
		"error: pool with 3 assets, must have two": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin, usdcCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        v16.USDCIBCDenom,
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

			// Validate CL and balancer pools have the same swap fee.
			suite.Require().Equal(balancerPool.GetSwapFee(suite.Ctx), clPoolReturned.GetSwapFee(suite.Ctx))

			// Validate that CL and balancer pools have the same denoms
			balancerDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, balancerPool.GetId())
			suite.Require().NoError(err)

			concentratedDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, clPoolReturned.GetId())
			suite.Require().NoError(err)

			// Order between balancer and concentrated might differ
			// because balancer lexicographically orders denoms but CL does not.
			suite.Require().ElementsMatch(balancerDenoms, concentratedDenoms)
			suite.Require().Equal(tc.expectedDenoms, concentratedDenoms)
		})
	}
}

func (suite *ConcentratedUpgradeTestSuite) TestCreateCanonicalConcentratedLiuqidityPoolAndMigrationLink() {
	suite.Setup()

	lockableDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx)
	longestLockableDuration := lockableDurations[len(lockableDurations)-1]

	tests := map[string]struct {
		poolLiquidity              sdk.Coins
		cfmmPoolIdToLinkWith       uint64
		desiredDenom0              string
		expectedBalancerDenoms     []string
		expectedConcentratedDenoms []string
		setupInvalidDuraitons      bool
		expectError                error
	}{
		"success - denoms reordered relative to balancer": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			// lexicographically ordered
			expectedBalancerDenoms: []string{daiCoin.Denom, desiredDenom0Coin.Denom},
			// determined by desired denom 0
			expectedConcentratedDenoms: []string{desiredDenom0Coin.Denom, daiCoin.Denom},
			desiredDenom0:              desiredDenom0,
		},
		"success - denoms are not reordered relative to balancer": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			// lexicographically ordered
			expectedBalancerDenoms: []string{daiCoin.Denom, desiredDenom0Coin.Denom},
			// determined by desired denom 0
			expectedConcentratedDenoms: []string{daiCoin.Denom, desiredDenom0Coin.Denom},
			desiredDenom0:              daiCoin.Denom,
		},
		"error: invalid denom 0": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        v16.USDCIBCDenom,
			expectError:          v16.NoDesiredDenomInPoolError{v16.USDCIBCDenom},
		},
		"error: pool with 3 assets, must have two": {
			poolLiquidity:        sdk.NewCoins(desiredDenom0Coin, daiCoin, usdcCoin),
			cfmmPoolIdToLinkWith: validPoolId,
			desiredDenom0:        v16.USDCIBCDenom,
			expectError:          v16.ErrMustHaveTwoDenoms,
		},
		"error: invalid denom durations": {
			poolLiquidity:         sdk.NewCoins(desiredDenom0Coin, daiCoin),
			cfmmPoolIdToLinkWith:  validPoolId,
			desiredDenom0:         desiredDenom0,
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

			// Get balancer gauges.
			gaugeToRedirect, _ := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerPool.GetId(), longestLockableDuration)

			gaugeToNotRedeirect, _ := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, balancerId2, longestLockableDuration)

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

			concentratedDenoms, err := suite.App.PoolManagerKeeper.RouteGetPoolDenoms(suite.Ctx, clPoolInState.GetId())
			suite.Require().NoError(err)

			// This check does not guarantee order.
			suite.Require().ElementsMatch(balancerDenoms, concentratedDenoms)

			// Validate order of balancer denoms is lexicographically sorted.
			suite.Require().Equal(tc.expectedBalancerDenoms, balancerDenoms)

			// Validate order of concentrated pool denoms which might be different from balancer.
			suite.Require().Equal(tc.expectedConcentratedDenoms, concentratedDenoms)

			// Validate that CFMM gauge is linked to the new concentrated pool.
			concentratedPoolGaugeId, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, clPoolInState.GetId(), suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx).Duration)
			suite.Require().NoError(err)

			distrInfo := suite.App.PoolIncentivesKeeper.GetDistrInfo(suite.Ctx)
			suite.Require().Equal(distrInfo.Records[0].GaugeId, concentratedPoolGaugeId)

			// Validate that distribution record from another pool is not redirected.
			suite.Require().Equal(distrInfo.Records[1].GaugeId, gaugeToNotRedeirect)

			// Validate migration record.
			migrationInfo, err := suite.App.GAMMKeeper.GetAllMigrationInfo(suite.Ctx)
			suite.Require().NoError(err)
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
