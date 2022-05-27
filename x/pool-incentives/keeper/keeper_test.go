package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v8/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v8/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v8/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestCreateBalancerPoolGauges() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.Ctx)
	suite.Equal(3, len(lockableDurations))

	for i := 0; i < 3; i++ {
		poolId := suite.PrepareBalancerPool()
		pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.NoError(err)

		poolLpDenom := gammtypes.GetPoolShareDenom(pool.GetId())

		// Same amount of gauges as lockableDurations must be created for every pool created.
		gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[0])
		suite.NoError(err)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		suite.Equal(lockableDurations[0], gauge.DistributeTo.Duration)

		gaugeId, err = keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[1])
		suite.NoError(err)
		gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		suite.Equal(lockableDurations[1], gauge.DistributeTo.Duration)

		gaugeId, err = keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[2])
		suite.NoError(err)
		gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		suite.Equal(lockableDurations[2], gauge.DistributeTo.Duration)
	}
}
