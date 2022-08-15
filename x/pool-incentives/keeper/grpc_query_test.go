package keeper_test

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v11/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"
)

var (
	isPerpetual  = true
	notPerpetual = false
)

func (suite *KeeperTestSuite) TestGaugeIds() {
	suite.SetupTest()

	queryClient := suite.queryClient

	// Unexisted pool
	_, err := queryClient.GaugeIds(context.Background(), &types.QueryGaugeIdsRequest{
		PoolId: 1,
	})
	suite.Error(err)

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx)
	suite.Equal(3, len(lockableDurations))

	poolId := suite.PrepareBalancerPool()
	pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
	suite.NoError(err)

	res, err := queryClient.GaugeIds(context.Background(), &types.QueryGaugeIdsRequest{
		PoolId: poolId,
	})
	suite.NoError(err)
	suite.Equal(3, len(res.GaugeIdsWithDuration))
	suite.Equal(lockableDurations[0], res.GaugeIdsWithDuration[0].Duration)
	suite.Equal(lockableDurations[1], res.GaugeIdsWithDuration[1].Duration)
	suite.Equal(lockableDurations[2], res.GaugeIdsWithDuration[2].Duration)

	poolLpDenom := gammtypes.GetPoolShareDenom(pool.GetId())
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, res.GaugeIdsWithDuration[0].GaugeId)
	suite.NoError(err)
	suite.Equal(0, len(gauge.Coins))
	suite.Equal(true, gauge.IsPerpetual)
	suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
	suite.Equal(lockableDurations[0], gauge.DistributeTo.Duration)

	gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, res.GaugeIdsWithDuration[1].GaugeId)
	suite.NoError(err)
	suite.Equal(0, len(gauge.Coins))
	suite.Equal(true, gauge.IsPerpetual)
	suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
	suite.Equal(lockableDurations[1], gauge.DistributeTo.Duration)

	gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, res.GaugeIdsWithDuration[2].GaugeId)
	suite.NoError(err)
	suite.Equal(0, len(gauge.Coins))
	suite.Equal(true, gauge.IsPerpetual)
	suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
	suite.Equal(lockableDurations[2], gauge.DistributeTo.Duration)
}

func (suite *KeeperTestSuite) TestDistrInfo() {
	suite.SetupTest()

	queryClient := suite.queryClient

	res, err := queryClient.DistrInfo(context.Background(), &types.QueryDistrInfoRequest{})
	suite.NoError(err)

	suite.Equal("0", res.DistrInfo.TotalWeight.String())
	suite.Equal(0, len(res.DistrInfo.Records))
}

func (suite *KeeperTestSuite) TestDistrInfo2() {
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

	res, err := queryClient.DistrInfo(context.Background(), &types.QueryDistrInfoRequest{})
	suite.NoError(err)

	suite.Equal("600", res.DistrInfo.TotalWeight.String())
	suite.Equal(3, len(res.DistrInfo.Records))
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

func (suite *KeeperTestSuite) TestIncentivizedPools() {
	suite.SetupTest()

	queryClient := suite.queryClient

	res, err := queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
	suite.NoError(err)

	suite.Equal(0, len(res.IncentivizedPools))
}

func (suite *KeeperTestSuite) TestIncentivizedPools2() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper
	queryClient := suite.queryClient

	poolId := suite.PrepareBalancerPool()
	poolId2 := suite.PrepareBalancerPool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.Ctx)
	suite.Equal(3, len(lockableDurations))

	gauge1Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	gauge2Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[1])
	suite.NoError(err)

	gauge3Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[2])
	suite.NoError(err)

	gauge4Id, err := keeper.GetPoolGaugeId(suite.Ctx, poolId2, lockableDurations[2])
	suite.NoError(err)

	// Create 4 records
	err = keeper.UpdateDistrRecords(suite.Ctx, types.DistrRecord{
		GaugeId: gauge1Id,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gauge2Id,
		Weight:  sdk.NewInt(200),
	}, types.DistrRecord{
		GaugeId: gauge3Id,
		Weight:  sdk.NewInt(300),
	}, types.DistrRecord{
		GaugeId: gauge4Id,
		Weight:  sdk.NewInt(300),
	})
	suite.NoError(err)

	res, err := queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
	suite.NoError(err)

	suite.Equal(4, len(res.IncentivizedPools))

	suite.Equal(poolId, res.IncentivizedPools[0].PoolId)
	suite.Equal(gauge1Id, res.IncentivizedPools[0].GaugeId)
	suite.Equal(time.Hour, res.IncentivizedPools[0].LockableDuration)

	suite.Equal(poolId, res.IncentivizedPools[1].PoolId)
	suite.Equal(gauge2Id, res.IncentivizedPools[1].GaugeId)
	suite.Equal(time.Hour*3, res.IncentivizedPools[1].LockableDuration)

	suite.Equal(poolId, res.IncentivizedPools[2].PoolId)
	suite.Equal(gauge3Id, res.IncentivizedPools[2].GaugeId)
	suite.Equal(time.Hour*7, res.IncentivizedPools[2].LockableDuration)

	suite.Equal(poolId2, res.IncentivizedPools[3].PoolId)
	suite.Equal(gauge4Id, res.IncentivizedPools[3].GaugeId)
	suite.Equal(time.Hour*7, res.IncentivizedPools[3].LockableDuration)

	// Actually, the pool incentives module can add incentives to any perpetual gauge, even if the gauge is not directly related to a pool.
	// However, these records must be excluded in incentivizedPools.
	gauge5Id, err := suite.App.IncentivesKeeper.CreateGauge(
		suite.Ctx, isPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "stake",
			Duration:      time.Hour,
		}, time.Now(), 1)
	suite.NoError(err)

	err = keeper.UpdateDistrRecords(suite.Ctx, types.DistrRecord{
		GaugeId: gauge1Id,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gauge2Id,
		Weight:  sdk.NewInt(200),
	}, types.DistrRecord{
		GaugeId: gauge3Id,
		Weight:  sdk.NewInt(300),
	}, types.DistrRecord{
		GaugeId: gauge4Id,
		Weight:  sdk.NewInt(300),
	}, types.DistrRecord{
		GaugeId: gauge5Id,
		Weight:  sdk.NewInt(300),
	})
	suite.NoError(err)

	res, err = queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
	suite.NoError(err)

	suite.Equal(4, len(res.IncentivizedPools))

	suite.Equal(poolId, res.IncentivizedPools[0].PoolId)
	suite.Equal(gauge1Id, res.IncentivizedPools[0].GaugeId)
	suite.Equal(time.Hour, res.IncentivizedPools[0].LockableDuration)

	suite.Equal(poolId, res.IncentivizedPools[1].PoolId)
	suite.Equal(gauge2Id, res.IncentivizedPools[1].GaugeId)
	suite.Equal(time.Hour*3, res.IncentivizedPools[1].LockableDuration)

	suite.Equal(poolId, res.IncentivizedPools[2].PoolId)
	suite.Equal(gauge3Id, res.IncentivizedPools[2].GaugeId)
	suite.Equal(time.Hour*7, res.IncentivizedPools[2].LockableDuration)

	suite.Equal(poolId2, res.IncentivizedPools[3].PoolId)
	suite.Equal(gauge4Id, res.IncentivizedPools[3].GaugeId)
	suite.Equal(time.Hour*7, res.IncentivizedPools[3].LockableDuration)

	// Ensure that non-perpetual pot can't get rewards.
	// TODO: extract this to standalone test
	gauge6Id, err := suite.App.IncentivesKeeper.CreateGauge(
		suite.Ctx, notPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "stake",
			Duration:      time.Hour,
		}, time.Now(), 1)

	suite.NoError(err)
	err = keeper.UpdateDistrRecords(suite.Ctx, types.DistrRecord{
		GaugeId: gauge6Id,
		Weight:  sdk.NewInt(100),
	})
	suite.Error(err)
}
