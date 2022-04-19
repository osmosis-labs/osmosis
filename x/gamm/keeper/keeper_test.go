package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/apptesting"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

var (
	acc1 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.SetupTestApp()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func (suite *KeeperTestSuite) prepareBalancerPoolWithPoolParams(poolParams balancer.PoolParams) uint64 {
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc, sdk.NewCoins(
			sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
			sdk.NewCoin("foo", sdk.NewInt(10000000)),
			sdk.NewCoin("bar", sdk.NewInt(10000000)),
			sdk.NewCoin("baz", sdk.NewInt(10000000)),
		))
		suite.Require().NoError(err)
	}

	poolAssets := []balancertypes.PoolAsset{
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

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(acc1, poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) prepareCustomBalancerPool(
	balances sdk.Coins,
	poolAssets []balancertypes.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, acc, balances)
		suite.Require().NoError(err)
	}

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(acc1, poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) prepareBalancerPool() uint64 {
	poolId := suite.prepareBalancerPoolWithPoolParams(balancer.PoolParams{
		SwapFee: sdk.NewDec(0),
		ExitFee: sdk.NewDec(0),
	})

	spotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "foo", "bar")
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewDec(2).String(), spotPrice.String())

	spotPrice, err = suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "bar", "baz")
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewDecWithPrec(15, 1).String(), spotPrice.String())

	spotPrice, err = suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, "baz", "foo")
	suite.Require().NoError(err)

	s := sdk.NewDec(1).Quo(sdk.NewDec(3))
	sp := s.Mul(types.SigFigs).RoundInt().ToDec().Quo(types.SigFigs)
	suite.Require().Equal(sp.String(), spotPrice.String())

	return poolId
}
