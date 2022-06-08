package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v9/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v9/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v9/x/gamm/types"
)

func (suite *KeeperTestSuite) TestJoinPoolExitPool() {
	suite.SetupTest()

	// Mint some assets to the accounts.
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, defaultAcctFunds)
	}

	createPoolAcc := suite.TestAccs[0]
	joinPoolAcc := suite.TestAccs[1]

	// Create a pool of "foo" and "bar"
	msgCreatePool := balancer.NewMsgCreateBalancerPool(createPoolAcc, balancer.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}, defaultPoolAssets, defaultFutureGovernor)
	poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msgCreatePool)
	suite.Require().NoError(err)

	msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)
	ctx := sdk.WrapSDKContext(suite.Ctx)

	msgJoinPool := types.MsgJoinPool{
		Sender: joinPoolAcc.String(),
		PoolId: poolId,
		ShareOutAmount: minShareOutAmount,
		TokenInMaxs: sdk.Coins{},
	}
	balancesBefore := suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc)

	_, err = msgServer.JoinPool(ctx, &msgJoinPool)
	suite.Require().NoError(err)

	balancesAfterJoin := suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc)
	deltaBalances, _ := balancesAfterJoin.SafeSub(balancesBefore)
	gammShareAmount := deltaBalances.AmountOf("gamm/pool/1")
	fooInAmount := deltaBalances.AmountOf("foo")
	barInAmount := deltaBalances.AmountOf("bar")
	
	suite.Require().Equal("-5000", fooInAmount.String())
	suite.Require().Equal("-5000", barInAmount.String())
	suite.Require().Equal("50000000000000000000", gammShareAmount.String())

	// now we test if exit pool returns same amount used in joinPool
	msgExitPool := types.MsgExitPool{
		Sender:        joinPoolAcc.String(),
		PoolId:        poolId,
		ShareInAmount: gammShareAmount,
		TokenOutMins:  sdk.Coins{},
	}
	_, err = msgServer.ExitPool(ctx, &msgExitPool)
	suite.Require().NoError(err)

	balancesAfterExit := suite.App.BankKeeper.GetAllBalances(suite.Ctx, joinPoolAcc)
	deltaBalances, _ = balancesAfterExit.SafeSub(balancesAfterJoin)
	gammShareAmount = deltaBalances.AmountOf("gamm/pool/1")
	fooInAmount = deltaBalances.AmountOf("foo")
	barInAmount = deltaBalances.AmountOf("bar")

	suite.Require().Equal("4950", fooInAmount.String())
	suite.Require().Equal("4950", barInAmount.String())
	suite.Require().Equal("-50000000000000000000", gammShareAmount.String())
}
