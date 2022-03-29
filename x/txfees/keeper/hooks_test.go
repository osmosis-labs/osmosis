package keeper_test

import (
	// "time"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (suite *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	suite.SetupTest(false)
	baseDenom, _ := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)

	testCases := []struct {
		name             string
	}{
		{
			name: "tc name",
		},
	}

	for _, tc := range testCases {

		fmt.Print("\n", tc)

		// we create three pools for three separate fee tokens
		uion := "uion"
		uionPoolId := suite.PreparePoolWithAssets(
			sdk.NewInt64Coin(baseDenom, 500),
			sdk.NewInt64Coin(uion, 500),
		)
		suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)

		atom := "atom"
		atomPoolId := suite.PreparePoolWithAssets(
			sdk.NewInt64Coin(baseDenom, 500),
			sdk.NewInt64Coin(atom, 500),
		)
		suite.ExecuteUpgradeFeeTokenProposal(atom, atomPoolId)

		ust := "ust"
		ustPoolId := suite.PreparePoolWithAssets(
			sdk.NewInt64Coin(baseDenom, 500),
			sdk.NewInt64Coin(ust, 500),
		)
		suite.ExecuteUpgradeFeeTokenProposal(ust, ustPoolId)

		coins := sdk.NewCoins(
			sdk.NewInt64Coin(uion, 10),
			sdk.NewInt64Coin(atom, 20),
			sdk.NewInt64Coin(ust, 14),
		)

		_, _, addr0 := testdata.KeyTestPubAddr()
		simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr0, coins)
		suite.app.BankKeeper.SendCoinsFromAccountToModule(suite.ctx, addr0, types.FooCollectorName, coins)

		moduleAddrFee := suite.app.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
		moduleAddrFoo := suite.app.AccountKeeper.GetModuleAddress(types.FooCollectorName)

		fmt.Print("\n(pre) Main module acc balances: ", suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrFee))
		fmt.Print("\n(pre) Second module acc balances: ", suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrFoo))

		// make sure module account is funded with test fee tokens
		suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrFoo, coins[0]))
		suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrFoo, coins[1]))
		suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrFoo, coins[2]))

		params := suite.app.IncentivesKeeper.GetParams(suite.ctx)
		futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Minute))
		suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))

		suite.Require().Empty(suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrFoo)) 

		fmt.Print("\n(post) Main module acc balances: ", suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrFee))
		fmt.Print("\n(post) Second module acc balances: ", suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrFoo))
	}
}
