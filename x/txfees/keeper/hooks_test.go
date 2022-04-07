package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (suite *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	suite.SetupTest(false)
	baseDenom, _ := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)

	// create pools for three separate fee tokens
	
	defaultPooledAssetAmount := int64(500)

	uion := "uion"
	uionPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(uion, defaultPooledAssetAmount),
	)
	uionPoolI, err:= suite.app.GAMMKeeper.GetPool(suite.ctx, uionPoolId)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)

	atom := "atom"
	atomPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(atom, defaultPooledAssetAmount),
	)
	atomPoolI, err:= suite.app.GAMMKeeper.GetPool(suite.ctx, atomPoolId)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(atom, atomPoolId)

	ust := "ust"
	ustPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(ust, defaultPooledAssetAmount),
	)
	ustPoolI, err:= suite.app.GAMMKeeper.GetPool(suite.ctx, ustPoolId)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(ust, ustPoolId)

	coins := sdk.NewCoins(sdk.NewInt64Coin(atom, 20),
		sdk.NewInt64Coin(atom, 20),
		sdk.NewInt64Coin(ust, 14))

	swapFee := sdk.NewDec(0)

	expectedOutput1 := uionPoolI.CalcOutAmtGivenIn(suite.ctx,
		sdk.NewCoins(sdk.NewInt64Coin(uion, 10)), 
		baseDenom, 
		swapFee).Amount.TruncateInt()
	expectedOutput2 := atomPoolI.CalcOutAmtGivenIn(suite.ctx,
		sdk.NewCoins(sdk.NewInt64Coin(atom, 20)), 
		baseDenom, 
		swapFee).Amount.TruncateInt()
	expectedOutput3 := ustPoolI.CalcOutAmtGivenIn(suite.ctx,
		sdk.NewCoins(sdk.NewInt64Coin(ust, 14)), 
		baseDenom, 
		swapFee).TruncateInt()
	
	fullExpectedOutput := expectedOutput1.Add(expectedOutput2).Add(expectedOutput3)

	_, _, addr0 := testdata.KeyTestPubAddr()
	simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr0, coins)
	suite.app.BankKeeper.SendCoinsFromAccountToModule(suite.ctx, addr0, types.NonNativeFeeCollectorName, coins)

	moduleAddrFee := suite.app.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	moduleAddrNonNativeFee := suite.app.AccountKeeper.GetModuleAddress(types.NonNativeFeeCollectorName)

	// make sure module account is funded with test fee tokens
	suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrNonNativeFee, coins[0]))
	suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrNonNativeFee, coins[1]))
	suite.Require().True(suite.app.BankKeeper.HasBalance(suite.ctx, moduleAddrNonNativeFee, coins[2]))

	params := suite.app.IncentivesKeeper.GetParams(suite.ctx)
	futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Minute))

	suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))

	suite.Require().Empty(suite.app.BankKeeper.GetAllBalances(suite.ctx, moduleAddrNonNativeFee)) 
	suite.Require().True(suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddrFee, baseDenom).Amount.GTE(fullExpectedOutput))
}
