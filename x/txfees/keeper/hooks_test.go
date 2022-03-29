package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gamm "github.com/osmosis-labs/osmosis/v7/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (suite *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	suite.SetupTest(false)
	baseDenom, _ := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)

	// create pools for three separate fee tokens
	
	defaultPooledAssetAmount := int64(500)

	uion := "uion"
	uionPoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(uion, defaultPooledAssetAmount),
	)
	suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)

	atom := "atom"
	atomPoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(atom, defaultPooledAssetAmount),
	)
	suite.ExecuteUpgradeFeeTokenProposal(atom, atomPoolId)

	ust := "ust"
	ustPoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(ust, defaultPooledAssetAmount),
	)
	suite.ExecuteUpgradeFeeTokenProposal(ust, ustPoolId)

	coins := sdk.NewCoins(
		sdk.NewInt64Coin(uion, 10),
		sdk.NewInt64Coin(atom, 20),
		sdk.NewInt64Coin(ust, 14),
	)

	gamm.CalcOutGivenIn(sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(coins[0].Amount), 
		sdk.NewDec(0),
	)

	expectedOutput1 := gamm.CalcOutGivenIn(sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(coins[0].Amount), 
		sdk.NewDec(0)).TruncateInt()
	expectedOutput2 := gamm.CalcOutGivenIn(sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(coins[1].Amount), 
		sdk.NewDec(0)).TruncateInt()
	expectedOutput3 := gamm.CalcOutGivenIn(sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(sdk.NewInt(defaultPooledAssetAmount)), 
		sdk.OneDec(), 
		sdk.NewDecFromInt(coins[2].Amount), 
		sdk.NewDec(0)).TruncateInt()
	
	fullExpectedOutput := expectedOutput1.Add(expectedOutput2).Add(expectedOutput3)

	_, _, addr0 := testdata.KeyTestPubAddr()
	simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr0, coins)
	suite.app.BankKeeper.SendCoinsFromAccountToModule(suite.ctx, addr0, types.FooCollectorName, coins)

	moduleAddrFee := suite.app.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	moduleAddrFoo := suite.app.AccountKeeper.GetModuleAddress(types.FooCollectorName)

	// make sure module account is funded with test fee tokens
	suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrFoo, coins[0]))
	suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrFoo, coins[1]))
	suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrFoo, coins[2]))

	params := suite.app.IncentivesKeeper.GetParams(suite.ctx)
	futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Minute))

	suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))

	suite.Require().Empty(suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrFoo)) 
	suite.Require().True(suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddrFee, baseDenom).Amount.GTE(fullExpectedOutput))
}
