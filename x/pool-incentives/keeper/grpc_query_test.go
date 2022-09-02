package keeper_test

import (
	"context"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

var (
	isPerpetual  = true
	notPerpetual = false
)

func (suite *KeeperTestSuite) TestGaugeIds() {
	for _, tc := range []struct {
		desc    string
		request *types.QueryGaugeIdsRequest
		err     bool
	}{
		{
			desc:    "Empty request",
			request: &types.QueryGaugeIdsRequest{},
			err:     true,
		},
		{
			desc: "Nonexistent pool",
			request: &types.QueryGaugeIdsRequest{
				PoolId: 2,
			},
			err: true,
		},
		{
			desc: "Happy case",
			request: &types.QueryGaugeIdsRequest{
				PoolId: 1,
			},
			err: false,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			queryClient := suite.queryClient
			// Prepare a balancer pool
			suite.PrepareBalancerPool()
			// LockableDurations should be 1, 3, 7 hours from the default genesis state.
			lockableDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx)
			suite.Require().Equal(3, len(lockableDurations))

			res, err := queryClient.GaugeIds(context.Background(), tc.request)
			if tc.err {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(3, len(res.GaugeIdsWithDuration))

				for i := 0; i < len(res.GaugeIdsWithDuration); i++ {
					suite.Require().Equal(lockableDurations[i], res.GaugeIdsWithDuration[i].Duration)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDistrInfo() {
	for _, tc := range []struct {
		desc                 string
		poolCreated          bool
		weights              []sdk.Int
		expectedTotalWeight  sdk.Int
		expectedRecordLength int
	}{
		{
			desc:                 "No pool exists",
			poolCreated:          false,
			weights:              []sdk.Int{},
			expectedTotalWeight:  sdk.NewInt(0),
			expectedRecordLength: 0,
		},
		{
			desc:                 "Happy case",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedTotalWeight:  sdk.NewInt(600),
			expectedRecordLength: 3,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			keeper := suite.App.PoolIncentivesKeeper
			queryClient := suite.queryClient

			if tc.poolCreated {
				poolId := suite.PrepareBalancerPool()

				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations := keeper.GetLockableDurations(suite.Ctx)
				suite.Require().Equal(3, len(lockableDurations))

				var distRecord []types.DistrRecord
				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[i])
					suite.Require().NoError(err)
					distRecord = append(distRecord, types.DistrRecord{
						GaugeId: gaugeId,
						Weight:  tc.weights[i],
					})
				}

				// Create 3 records
				err := keeper.UpdateDistrRecords(suite.Ctx, distRecord...)
				suite.Require().NoError(err)
			}

			res, err := queryClient.DistrInfo(context.Background(), &types.QueryDistrInfoRequest{})
			suite.Require().NoError(err)

			suite.Require().Equal(tc.expectedTotalWeight, res.DistrInfo.TotalWeight)
			suite.Require().Equal(tc.expectedRecordLength, len(res.DistrInfo.Records))
		})
	}
}

func (suite *KeeperTestSuite) TestParams() {
	suite.SetupTest()

	queryClient := suite.queryClient

	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	// Minted denom set as "stake" from the default genesis state
	suite.Require().Equal("stake", res.Params.MintedDenom)
}

func (suite *KeeperTestSuite) TestLockableDurations() {
	suite.SetupTest()

	queryClient := suite.queryClient

	res, err := queryClient.LockableDurations(context.Background(), &types.QueryLockableDurationsRequest{})
	suite.Require().NoError(err)

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	suite.Require().Equal(3, len(res.LockableDurations))
	suite.Require().Equal(time.Hour, res.LockableDurations[0])
	suite.Require().Equal(time.Hour*3, res.LockableDurations[1])
	suite.Require().Equal(time.Hour*7, res.LockableDurations[2])
}

func (suite *KeeperTestSuite) TestIncentivizedPools() {
	for _, tc := range []struct {
		desc                 string
		poolCreated          bool
		weights              []sdk.Int
		perpetual            bool
		nonPerpetual         bool
		expectedRecordLength int
	}{
		{
			desc:                 "No pool exist",
			poolCreated:          false,
			weights:              []sdk.Int{},
			expectedRecordLength: 0,
		},
		{
			desc:                 "Normal case",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedRecordLength: 3,
		},
		{
			desc:                 "Perpetual",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			perpetual:            true,
			expectedRecordLength: 3,
		},
		{
			desc:                 "Non Perpetual",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			nonPerpetual:         true,
			expectedRecordLength: 0,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			keeper := suite.App.PoolIncentivesKeeper
			queryClient := suite.queryClient
			var poolId uint64

			var lockableDurations []time.Duration
			if tc.poolCreated {
				// prepare a balancer pool
				poolId = suite.PrepareBalancerPool()
				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations = keeper.GetLockableDurations(suite.Ctx)
				suite.Require().Equal(3, len(lockableDurations))

				var distRecords []types.DistrRecord
				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[i])
					suite.Require().NoError(err)
					distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeId, Weight: tc.weights[i]})

					if tc.perpetual {
						gaugePerpetualId, err := suite.App.IncentivesKeeper.CreateGauge(
							suite.Ctx, isPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1)
						suite.Require().NoError(err)
						distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugePerpetualId, Weight: sdk.NewInt(300)})
					}
					if tc.nonPerpetual {
						gaugeNonPerpetualId, err := suite.App.IncentivesKeeper.CreateGauge(
							suite.Ctx, notPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1)
						suite.Require().NoError(err)
						distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeNonPerpetualId, Weight: sdk.NewInt(100)})
					}

					// Sort in ascending order of gaugeId
					sort.Slice(distRecords[:], func(i, j int) bool {
						return distRecords[i].GaugeId < distRecords[j].GaugeId
					})

					// update records and ensure that non-perpetuals pot cannot get rewards.
					keeper.UpdateDistrRecords(suite.Ctx, distRecords...)
				}
				res, err := queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedRecordLength, len(res.IncentivizedPools))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGaugeIncentivePercentage() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper
	queryClient := suite.queryClient

	poolId := suite.PrepareBalancerPool()
	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.Ctx)
	suite.Require().Equal(3, len(lockableDurations))

	gauge1Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[0])
	suite.Require().NoError(err)

	gauge2Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[1])
	suite.Require().NoError(err)

	gauge3Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[2])
	suite.Require().NoError(err)

	// Create 3 records
	err = keeper.UpdateDistrRecords(suite.Ctx, types.DistrRecord{
		GaugeId: gauge1Id,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gauge2Id,
		Weight:  sdk.NewInt(200),
	}, types.DistrRecord{
		GaugeId: gauge3Id,
		Weight:  sdk.NewInt(300),
	})
	suite.Require().NoError(err)

	res, err := queryClient.GaugeIds(context.Background(), &types.QueryGaugeIdsRequest{
		PoolId: poolId,
	})

	suite.Require().NoError(err)
	suite.Require().Equal("16.666666666666666700", res.GaugeIdsWithDuration[0].GaugeIncentivePercentage)
	suite.Require().Equal("33.333333333333333300", res.GaugeIdsWithDuration[1].GaugeIncentivePercentage)
	suite.Require().Equal("50.000000000000000000", res.GaugeIdsWithDuration[2].GaugeIncentivePercentage)
}
