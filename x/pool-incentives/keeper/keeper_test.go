package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.SetupTestApp()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

var (
	acc1 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

func (suite *KeeperTestSuite) prepareBalancerPoolWithPoolParams(PoolParams balancer.PoolParams) uint64 {
	// Mint some assets to the accounts.
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc, sdk.NewCoins(
			sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
			sdk.NewCoin("foo", sdk.NewInt(10000000)),
			sdk.NewCoin("bar", sdk.NewInt(10000000)),
			sdk.NewCoin("baz", sdk.NewInt(10000000)),
		))
		if err != nil {
			panic(err)
		}
	}

	poolAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(300),
			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
		},
	}
	msg := balancer.NewMsgCreateBalancerPool(acc1, PoolParams, poolAssets, "")
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
	suite.NoError(err)
	return poolId
}

func (suite *KeeperTestSuite) prepareBalancerPool() uint64 {
	poolId := suite.prepareBalancerPoolWithPoolParams(balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	})

	spotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "foo", "bar")
	suite.NoError(err)
	suite.Equal(sdk.NewDec(2).String(), spotPrice.String())
	spotPrice, err = suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "bar", "baz")
	suite.NoError(err)
	suite.Equal(sdk.NewDecWithPrec(15, 1).String(), spotPrice.String())
	spotPrice, err = suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "baz", "foo")
	suite.NoError(err)
	s := sdk.NewDec(1).Quo(sdk.NewDec(3))
	sp := s.Mul(gammtypes.SigFigs).RoundInt().ToDec().Quo(gammtypes.SigFigs)
	suite.Equal(sp.String(), spotPrice.String())

	return poolId
}

func (suite *KeeperTestSuite) TestCreateBalancerPoolGauges() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.Ctx)
	suite.Equal(3, len(lockableDurations))

	for i := 0; i < 3; i++ {
		poolId := suite.prepareBalancerPool()
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
