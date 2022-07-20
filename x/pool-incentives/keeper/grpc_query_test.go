package keeper_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
)

var (
	isPerpetual  = true
	notPerpetual = false
)

func (suite *KeeperTestSuite) TestGaugeIds() {
	suite.SetupTest()

	queryClient := suite.queryClient

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx)
	suite.Equal(3, len(lockableDurations))

	poolId := suite.PrepareBalancerPool()
	pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
	suite.NoError(err)

	for _, tc := range []struct {
		desc    string
		request *types.QueryGaugeIdsRequest
		err     bool
	}{
		{
			desc: "Empty request",
			request: &types.QueryGaugeIdsRequest{
			},
			err: true,
		},
		{
			desc: "Unexisted pool",
			request: &types.QueryGaugeIdsRequest{
				PoolId: 2,
			}, 
			err: true,
		},
		{
			desc: "Happy case",
			request: &types.QueryGaugeIdsRequest{
				PoolId: poolId,
			}, 
			err: false,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			res, err := queryClient.GaugeIds(context.Background(), tc.request)
			if tc.err {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.NoError(err)
				suite.Equal(3, len(res.GaugeIdsWithDuration))

				poolLpDenom := gammtypes.GetPoolShareDenom(pool.GetId())
				for i := 0; i < len(res.GaugeIdsWithDuration); i++ {
					suite.Equal(lockableDurations[i], res.GaugeIdsWithDuration[i].Duration)

					gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, res.GaugeIdsWithDuration[i].GaugeId)
					suite.NoError(err)
					suite.Equal(0, len(gauge.Coins))
					suite.Equal(true, gauge.IsPerpetual)
					suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
					suite.Equal(lockableDurations[i], gauge.DistributeTo.Duration)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDistrInfo() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper
	queryClient := suite.queryClient

	for _, tc := range []struct {
		desc    				string
		poolCreated 			bool
		weights      			[]sdk.Int
		expectedTotalWeight		sdk.Int
		expectedRecordLength	int
	}{
		{
			desc: "No pool exist",
			poolCreated: false,
			weights: []sdk.Int{},
			expectedTotalWeight: sdk.NewInt(0),
			expectedRecordLength: 0,
		},
		{
			desc: "Happy case",
			poolCreated: true,
			weights: []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedTotalWeight: sdk.NewInt(600),
			expectedRecordLength: 3,
		},
	}{
		tc := tc
		suite.Run(tc.desc, func() {
			if tc.poolCreated {
				poolId := suite.PrepareBalancerPool()

				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations := keeper.GetLockableDurations(suite.Ctx)
				suite.Equal(3, len(lockableDurations))

				var distRecord []types.DistrRecord
				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[i])
					suite.NoError(err)
					distRecord = append(distRecord, types.DistrRecord{
						GaugeId: gaugeId,
						Weight:  tc.weights[i],
					})
				}

				// Create 3 records
				err := keeper.UpdateDistrRecords(suite.Ctx, distRecord...)
				suite.NoError(err)
			}

			res, err := queryClient.DistrInfo(context.Background(), &types.QueryDistrInfoRequest{})
			suite.NoError(err)

			suite.Equal(tc.expectedTotalWeight, res.DistrInfo.TotalWeight)
			suite.Equal(tc.expectedRecordLength, len(res.DistrInfo.Records))
		})
	}
}

func (suite *KeeperTestSuite) TestParams() {
	suite.SetupTest()

	queryClient := suite.queryClient

	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.NoError(err)

	// Minted denom set as "stake" from the default genesis state
	suite.Equal("stake", res.Params.MintedDenom)
}

func (suite *KeeperTestSuite) TestLockableDurations() {
	suite.SetupTest()

	queryClient := suite.queryClient

	res, err := queryClient.LockableDurations(context.Background(), &types.QueryLockableDurationsRequest{})
	suite.NoError(err)

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	suite.Equal(3, len(res.LockableDurations))
	suite.Equal(time.Hour, res.LockableDurations[0])
	suite.Equal(time.Hour*3, res.LockableDurations[1])
	suite.Equal(time.Hour*7, res.LockableDurations[2])
}

func (suite *KeeperTestSuite) TestIncentivizedPools2() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper
	ammKeeper := suite.App.GAMMKeeper
	queryClient := suite.queryClient
	var poolId uint64

	gaugePerpetualId, err := suite.App.IncentivesKeeper.CreateGauge(
		suite.Ctx, isPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "stake",
			Duration:      time.Hour,
		}, time.Now(), 1)
	suite.NoError(err)

	gaugeNonPerpetualId, err := suite.App.IncentivesKeeper.CreateGauge(
		suite.Ctx, notPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "stake",
			Duration:      time.Hour,
		}, time.Now(), 1)

	suite.NoError(err)

	for _, tc := range []struct {
		desc    				string
		poolCreated 			bool
		distRecords      		[]types.DistrRecord
		expectedErr				bool
		expectedRecordLength	int
	}{
		{
			desc: "No pool exist",
			poolCreated: false,
			distRecords: []types.DistrRecord{},
			expectedErr: false,
			expectedRecordLength: 0,
		},
		{
			desc: "Normal case",
			poolCreated: true,
			distRecords: []types.DistrRecord{},
			expectedErr: false,
			expectedRecordLength: 3,
		},
		{
			desc: "Perpetual",
			poolCreated: true,
			distRecords: []types.DistrRecord{types.DistrRecord{
				GaugeId: gaugePerpetualId,
				Weight:  sdk.NewInt(300),
			}},
			expectedErr: false,
			expectedRecordLength: 3,
		},
		{
			desc: "Non Perpetual",
			poolCreated: true,
			distRecords: []types.DistrRecord{types.DistrRecord{
				GaugeId: gaugeNonPerpetualId,
				Weight:  sdk.NewInt(100),
			}},
			expectedErr: true,
			expectedRecordLength: 0,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			var lockableDurations []time.Duration
			if tc.poolCreated {
				// only create a pool
				pool, err := ammKeeper.GetPoolsAndPoke(suite.Ctx)
				if len(pool) == 0 {
					poolId = suite.PrepareBalancerPool()
				}				
				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations = keeper.GetLockableDurations(suite.Ctx)
				suite.Equal(3, len(lockableDurations))

				var distRecords []types.DistrRecord
				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[i])
					suite.NoError(err)

					distRecords = append(distRecords, types.DistrRecord{
						GaugeId: gaugeId,
						Weight:  sdk.NewInt(100),
					})
				}
				tc.distRecords = append(distRecords, tc.distRecords...)

				distRecords = tc.distRecords
				sort.Slice(distRecords[:], func(i, j int) bool {
					return distRecords[i].GaugeId < distRecords[j].GaugeId
				})
				fmt.Println(distRecords)
				// Create records
				err = keeper.UpdateDistrRecords(suite.Ctx, distRecords...)
				if tc.expectedErr {
					suite.Require().Error(err)
				} else {
					suite.Require().NoError(err)
				}
			}
			res, err := queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
			if !tc.expectedErr {
				suite.NoError(err)
				suite.Equal(tc.expectedRecordLength, len(res.IncentivizedPools))

				for i := 0; i < 0; i++ {
					if tc.poolCreated {
						suite.Equal(poolId, res.IncentivizedPools[i].PoolId)
					}
					suite.Equal(tc.distRecords[i].GaugeId, res.IncentivizedPools[i].GaugeId)
					suite.Equal(lockableDurations[i], res.IncentivizedPools[i].LockableDuration)
				}
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
	suite.Equal(3, len(lockableDurations))

	gauge1Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	gauge2Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[1])
	suite.NoError(err)

	gauge3Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[2])
	suite.NoError(err)

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
	suite.NoError(err)

	res, err := queryClient.GaugeIds(context.Background(), &types.QueryGaugeIdsRequest{
		PoolId: poolId,
	})

	suite.NoError(err)
	suite.Equal("16.666666666666666700", res.GaugeIdsWithDuration[0].GaugeIncentivePercentage)
	suite.Equal("33.333333333333333300", res.GaugeIdsWithDuration[1].GaugeIncentivePercentage)
	suite.Equal("50.000000000000000000", res.GaugeIdsWithDuration[2].GaugeIncentivePercentage)
}
