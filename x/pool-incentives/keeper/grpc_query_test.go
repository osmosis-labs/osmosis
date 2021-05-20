package keeper_test

import (
	"context"

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
