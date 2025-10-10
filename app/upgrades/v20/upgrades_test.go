package v20_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	v20 "github.com/osmosis-labs/osmosis/v31/app/upgrades/v20"
	gammmigration "github.com/osmosis-labs/osmosis/v31/x/gamm/types/migration"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v31/x/pool-incentives/types"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// Validates that the pool incentive distribution records are updated correctly.
// The expected update is to convert concentrated gauges to group gauges iff
// - migration record exists for the concentrated pool and another CFMM pool
// - if migration between concentrated and CFMM exists, then the CFMM pool is not incentivized individually
// It is expected that all other distribution records are not modified.
func (s *UpgradeTestSuite) TestCreateGroupsForIncentivePairs() {
	s.SetupTest()

	// Create pools once for all tests
	poolInfo := s.PrepareAllSupportedPools()
	secondPoolInfo := s.PrepareAllSupportedPools()
	thirdPoolInfo := s.PrepareAllSupportedPools()

	var (
		expectedGroupGaugeID  = s.App.IncentivesKeeper.GetLastGaugeID(s.Ctx) + 1
		defaultWeight         = osmomath.NewInt(10)
		groupGaugeDistrRecord = []poolincentivestypes.DistrRecord{
			{
				GaugeId: expectedGroupGaugeID,
				Weight:  defaultWeight,
			},
		}
		concentratedDistRecord = []poolincentivestypes.DistrRecord{
			{
				GaugeId: poolInfo.ConcentratedGaugeID,
				Weight:  defaultWeight,
			},
		}
		balancerDistrRecord = []poolincentivestypes.DistrRecord{
			{
				GaugeId: poolInfo.BalancerGaugeID,
				Weight:  defaultWeight,
			},
		}
		stableswapDistrRecord = []poolincentivestypes.DistrRecord{
			{
				GaugeId: poolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
		}

		noMigrationInfo = []gammmigration.BalancerToConcentratedPoolLink{}
	)

	// Test for individual distribution record configurations.
	tests := map[string]struct {
		migrationInfo               []gammmigration.BalancerToConcentratedPoolLink
		distributionRecords         []poolincentivestypes.DistrRecord
		expectError                 error
		expectedDistributionRecords []poolincentivestypes.DistrRecord
	}{
		"one distr record with concentrated pool linked to gamm (converted to group)": {
			migrationInfo: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: poolInfo.BalancerPoolID,
					ClPoolId:       poolInfo.ConcentratedPoolID,
				},
			},
			distributionRecords: concentratedDistRecord,

			expectedDistributionRecords: groupGaugeDistrRecord,
		},
		"one distr record with balancer pool (no-op)": {
			// No migration link
			migrationInfo:       noMigrationInfo,
			distributionRecords: balancerDistrRecord,

			expectedDistributionRecords: balancerDistrRecord,
		},
		"one distr record with stableswap pool (no-op)": {
			// No migration link
			migrationInfo:       noMigrationInfo,
			distributionRecords: stableswapDistrRecord,

			expectedDistributionRecords: stableswapDistrRecord,
		},
		"one distr record with concentrated pool that is not linked to gamm (no-op)": {
			// No migration link
			migrationInfo: noMigrationInfo,

			distributionRecords: concentratedDistRecord,

			expectedDistributionRecords: concentratedDistRecord,
		},
		"migration link between gamm pool and concentrated pool exists but no distr record (no-op)": {
			migrationInfo: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: poolInfo.BalancerPoolID,
					ClPoolId:       poolInfo.ConcentratedPoolID,
				},
			},

			distributionRecords: []poolincentivestypes.DistrRecord{},

			expectedDistributionRecords: []poolincentivestypes.DistrRecord(nil),
		},
		"migration link between gamm pool and concentrated pool exists; no distr record with community pool gauge ID (no-op)": {
			migrationInfo: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: poolInfo.BalancerPoolID,
					ClPoolId:       poolInfo.ConcentratedPoolID,
				},
			},

			distributionRecords: []poolincentivestypes.DistrRecord{{
				GaugeId: poolincentivestypes.CommunityPoolDistributionGaugeID,
				Weight:  defaultWeight,
			}},

			expectedDistributionRecords: []poolincentivestypes.DistrRecord{{
				GaugeId: poolincentivestypes.CommunityPoolDistributionGaugeID,
				Weight:  defaultWeight,
			}},
		},

		// error cases

		"error: one distr record with balancer pool but migration link present": {
			migrationInfo: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: poolInfo.BalancerPoolID,
					ClPoolId:       poolInfo.ConcentratedPoolID,
				},
			},
			distributionRecords: balancerDistrRecord,

			expectError: v20.IncentivizedCFMMDirectWhenMigrationLinkPresentError{
				CFMMPoolID:         poolInfo.BalancerPoolID,
				ConcentratedPoolID: poolInfo.ConcentratedPoolID,
				CFMMGaugeID:        poolInfo.BalancerGaugeID,
			},
		},
		// This is an invalid setup since we do not support stableswap in migration
		// link but we test it for completeness.
		"error: one distr record with stableswap pool but migration link present": {
			migrationInfo: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: poolInfo.StableSwapPoolID,
					ClPoolId:       poolInfo.ConcentratedPoolID,
				},
			},
			distributionRecords: stableswapDistrRecord,

			expectError: v20.IncentivizedCFMMDirectWhenMigrationLinkPresentError{
				CFMMPoolID:         poolInfo.StableSwapPoolID,
				ConcentratedPoolID: poolInfo.ConcentratedPoolID,
				CFMMGaugeID:        poolInfo.StableSwapGaugeID,
			},
		},
		// expected to create a group for the concentrated pool
		// however, detect that there is a balancer pool that is incentivized individually
		// and fail.
		"error: one distr record with balancer pool and concentrated pool but migration link present": {
			migrationInfo: []gammmigration.BalancerToConcentratedPoolLink{
				{
					BalancerPoolId: poolInfo.BalancerPoolID,
					ClPoolId:       poolInfo.ConcentratedPoolID,
				},
			},
			distributionRecords: append(concentratedDistRecord, balancerDistrRecord...),

			expectError: v20.IncentivizedCFMMDirectWhenMigrationLinkPresentError{
				CFMMPoolID:         poolInfo.BalancerPoolID,
				ConcentratedPoolID: poolInfo.ConcentratedPoolID,
				CFMMGaugeID:        poolInfo.BalancerGaugeID,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			s.runCreateGroupsForIncentivePairsTest(tc.migrationInfo, tc.distributionRecords, tc.expectedDistributionRecords, expectedGroupGaugeID, tc.expectError)
		})
	}

	// 2 concentrated pools linked to 2 balancer pools - converted to group
	// linked balancer pools are not incentivized individually
	// separate individual balancer pool - no-op
	// 2 separate individual stableswap pools - no-op
	// concentrated pool that does not have migration link - no-op
	s.Run("valid multi distr record test", func() {

		// Refetch updated expected group gauge ID
		// since previous tests might have modified it by creating their own groups
		// during tests
		var expectedGroupGaugeID = s.App.IncentivesKeeper.GetLastGaugeID(s.Ctx) + 1

		// Configure migration records only for 2 groups of concentrated and balancer pools.
		// Note that the third one does not have a migration record.
		migrationInfo := []gammmigration.BalancerToConcentratedPoolLink{
			{
				BalancerPoolId: poolInfo.BalancerPoolID,
				ClPoolId:       poolInfo.ConcentratedPoolID,
			},
			{
				BalancerPoolId: secondPoolInfo.BalancerPoolID,
				ClPoolId:       secondPoolInfo.ConcentratedPoolID,
			},
		}

		distributionRecords := []poolincentivestypes.DistrRecord{
			{
				GaugeId: poolInfo.ConcentratedGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: poolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: secondPoolInfo.ConcentratedGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: secondPoolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: thirdPoolInfo.ConcentratedGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: thirdPoolInfo.BalancerGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: thirdPoolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
		}

		expectedDistributionRecords := []poolincentivestypes.DistrRecord{
			{
				// Replaced with group gauge
				GaugeId: expectedGroupGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: poolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
			{
				// Replaced with group gauge
				GaugeId: expectedGroupGaugeID + 1,
				Weight:  defaultWeight,
			},
			{
				GaugeId: secondPoolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
			{
				// Not replaced with group gauge because
				// migration record does not exist for this pool and another balancer pool
				GaugeId: thirdPoolInfo.ConcentratedGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: thirdPoolInfo.BalancerGaugeID,
				Weight:  defaultWeight,
			},
			{
				GaugeId: thirdPoolInfo.StableSwapGaugeID,
				Weight:  defaultWeight,
			},
		}

		noError := error(nil)
		s.runCreateGroupsForIncentivePairsTest(migrationInfo, distributionRecords, expectedDistributionRecords, expectedGroupGaugeID, noError)
	})
}

func (s *UpgradeTestSuite) runCreateGroupsForIncentivePairsTest(migrationInfo []gammmigration.BalancerToConcentratedPoolLink, distributionRecords []poolincentivestypes.DistrRecord, expectedDistributionRecords []poolincentivestypes.DistrRecord, expectedGroupGaugeID uint64, expectedError error) {
	// Configure migration records for each test individually (overwrites previous migration records).
	err := s.App.GAMMKeeper.OverwriteMigrationRecords(s.Ctx, gammmigration.MigrationRecords{BalancerToConcentratedPoolLinks: migrationInfo})
	s.Require().NoError(err)

	// Configure distribution records for each test individually (overwrites previous distribution records).
	err = s.App.PoolIncentivesKeeper.ReplaceDistrRecords(s.Ctx, distributionRecords...)
	s.Require().NoError(err)

	cacheCtx, write := s.Ctx.CacheContext()
	err = v20.CreateGroupsForIncentivePairs(cacheCtx, &s.App.AppKeepers)
	if expectedError != nil {
		s.Require().Error(err)
		s.Require().ErrorIs(expectedError, err)
		return
	}

	// Only write cache context on success since our test cases depend on each other.
	// Persisting it after errors will cause the next test to fail.
	write()

	s.Require().NoError(err)

	// Validate that final distribution records are as expected
	updatedDistrInfo := s.App.PoolIncentivesKeeper.GetDistrInfo(s.Ctx)
	s.Require().Equal(expectedDistributionRecords, updatedDistrInfo.Records)

	// Validate that the group gauge along with the group were created if applicable.
	for _, expectedRecord := range expectedDistributionRecords {
		if expectedRecord.GaugeId == expectedGroupGaugeID {
			s.ValidateGroupExists(expectedGroupGaugeID)
		}
	}
}
