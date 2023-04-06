package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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

func (suite *KeeperTestSuite) TestCreateConcentratePoolGauges() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper

	for i := 0; i < 3; i++ {
		clPool := suite.PrepareConcentratedPool()

		incParams := suite.App.IncentivesKeeper.GetParams(suite.Ctx).DistrEpochIdentifier
		currEpoch := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, incParams)

		// Same amount of gauges as lockableDurations must be created for every pool created.
		gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, clPool.GetId(), currEpoch.Duration)
		suite.NoError(err)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(gaugeId, gauge.Id)
	}
}

func (suite *KeeperTestSuite) TestCreatePoolGauges() {
	tests := []struct {
		name        string
		poolId      uint64
		poolType    poolmanagertypes.PoolType
		expectedErr bool
	}{
		{
			name:        "Concentrated Liquidity Pool",
			poolId:      1,
			poolType:    poolmanagertypes.Concentrated,
			expectedErr: false,
		},
		{
			name:        "non concentrated pool",
			poolId:      2,
			poolType:    poolmanagertypes.Balancer,
			expectedErr: false,
		},
		{
			name:        "non existent pool",
			poolId:      0,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.PrepareConcentratedPool().GetId()
			suite.PrepareBalancerPool()

			var err error
			// TODO: split into separate tests
			if tc.poolType == poolmanagertypes.Concentrated {
				err = suite.App.PoolIncentivesKeeper.CreateConcentratedLiquidityPoolGauge(suite.Ctx, tc.poolId)
			} else {
				err = suite.App.PoolIncentivesKeeper.CreateLockablePoolGauges(suite.Ctx, tc.poolId)
			}

			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				var lockableDuration time.Duration
				if tc.poolType == poolmanagertypes.Concentrated {
					epochInfo := suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx)
					lockableDuration = epochInfo.Duration
				} else {
					lockableDuration = time.Hour * 7
				}

				// make sure gauge is created and check that gaugeId is associated with poolId
				_, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, tc.poolId, lockableDuration)
				suite.Require().NoError(err)
			}
		})
	}
}
