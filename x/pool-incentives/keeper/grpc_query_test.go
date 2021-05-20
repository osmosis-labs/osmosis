package keeper_test

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
)

func (suite *KeeperTestSuite) TestPotIds() {
	suite.SetupTest()

	queryClient := suite.queryClient

	// Unexisted pool
	_, err := queryClient.PotIds(context.Background(), &types.QueryPotIdsRequest{
		PoolId: 1,
	})
	suite.Error(err)

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := suite.app.PoolIncentivesKeeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	poolId := suite.preparePool()
	pool, err := suite.app.GAMMKeeper.GetPool(suite.ctx, poolId)
	suite.NoError(err)

	res, err := queryClient.PotIds(context.Background(), &types.QueryPotIdsRequest{
		PoolId: poolId,
	})
	suite.NoError(err)
	suite.Equal(3, len(res.PotIdsWithDuration))
	suite.Equal(lockableDurations[0], res.PotIdsWithDuration[0].Duration)
	suite.Equal(lockableDurations[1], res.PotIdsWithDuration[1].Duration)
	suite.Equal(lockableDurations[2], res.PotIdsWithDuration[2].Duration)

	pot, err := suite.app.IncentivesKeeper.GetPotByID(suite.ctx, res.PotIdsWithDuration[0].PotId)
	suite.NoError(err)
	suite.Equal(0, len(pot.Coins))
	suite.Equal(true, pot.IsPerpetual)
	suite.Equal(pool.GetTotalShare().Denom, pot.DistributeTo.Denom)
	suite.Equal(lockableDurations[0], pot.DistributeTo.Duration)

	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, res.PotIdsWithDuration[1].PotId)
	suite.NoError(err)
	suite.Equal(0, len(pot.Coins))
	suite.Equal(true, pot.IsPerpetual)
	suite.Equal(pool.GetTotalShare().Denom, pot.DistributeTo.Denom)
	suite.Equal(lockableDurations[1], pot.DistributeTo.Duration)

	pot, err = suite.app.IncentivesKeeper.GetPotByID(suite.ctx, res.PotIdsWithDuration[2].PotId)
	suite.NoError(err)
	suite.Equal(0, len(pot.Coins))
	suite.Equal(true, pot.IsPerpetual)
	suite.Equal(pool.GetTotalShare().Denom, pot.DistributeTo.Denom)
	suite.Equal(lockableDurations[2], pot.DistributeTo.Duration)
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

	keeper := suite.app.PoolIncentivesKeeper
	queryClient := suite.queryClient

	poolId := suite.preparePool()

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.ctx)
	suite.Equal(3, len(lockableDurations))

	pot1Id, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[0])
	suite.NoError(err)

	pot2Id, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[1])
	suite.NoError(err)

	pot3Id, err := keeper.GetPoolPotId(suite.ctx, poolId, lockableDurations[2])
	suite.NoError(err)

	// Create 3 records
	err = keeper.UpdateDistrRecords(suite.ctx, types.DistrRecord{
		PotId:  pot1Id,
		Weight: sdk.NewInt(100),
	}, types.DistrRecord{
		PotId:  pot2Id,
		Weight: sdk.NewInt(200),
	}, types.DistrRecord{
		PotId:  pot3Id,
		Weight: sdk.NewInt(300),
	})
	suite.NoError(err)

	// Unexisted pool
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
