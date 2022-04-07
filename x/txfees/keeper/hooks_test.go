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
	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)

	// create pools for three separate fee tokens

	defaultPooledAssetAmount := int64(500)

	uion := "uion"
	uionPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(uion, defaultPooledAssetAmount),
	)
	uionPoolI, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, uionPoolId)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)

	atom := "atom"
	atomPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(atom, defaultPooledAssetAmount),
	)
	atomPoolI, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, atomPoolId)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(atom, atomPoolId)

	ust := "ust"
	ustPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(ust, defaultPooledAssetAmount),
	)
	ustPoolI, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, ustPoolId)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(ust, ustPoolId)

	coins := sdk.NewCoins(sdk.NewInt64Coin(atom, 20),
		sdk.NewInt64Coin(atom, 20),
		sdk.NewInt64Coin(ust, 14))

	swapFee := sdk.NewDec(0)

	expectedOutput1, err := uionPoolI.CalcOutAmtGivenIn(suite.Ctx,
		sdk.NewCoins(sdk.NewInt64Coin(uion, 10)),
		baseDenom,
		swapFee)
	expectedOutput2, err := atomPoolI.CalcOutAmtGivenIn(suite.Ctx,
		sdk.NewCoins(sdk.NewInt64Coin(atom, 20)),
		baseDenom,
		swapFee)
	expectedOutput3, err := ustPoolI.CalcOutAmtGivenIn(suite.Ctx,
		sdk.NewCoins(sdk.NewInt64Coin(ust, 14)),
		baseDenom,
		swapFee)

	fullExpectedOutput := expectedOutput1.Add(expectedOutput2).Add(expectedOutput3)

	_, _, addr0 := testdata.KeyTestPubAddr()
	simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, coins)
	suite.App.BankKeeper.SendCoinsFromAccountToModule(suite.Ctx, addr0, types.NonNativeFeeCollectorName, coins)

	moduleAddrFee := suite.App.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	moduleAddrNonNativeFee := suite.App.AccountKeeper.GetModuleAddress(types.NonNativeFeeCollectorName)

	// make sure module account is funded with test fee tokens
	suite.Require().True(suite.App.BankKeeper.HasBalance(suite.Ctx, moduleAddrNonNativeFee, coins[0]))
	suite.Require().True(suite.App.BankKeeper.HasBalance(suite.Ctx, moduleAddrNonNativeFee, coins[1]))
	suite.Require().True(suite.App.BankKeeper.HasBalance(suite.Ctx, moduleAddrNonNativeFee, coins[2]))

	params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	futureCtx := suite.Ctx.WithBlockTime(time.Now().Add(time.Minute))

	suite.App.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))

	suite.Require().Empty(suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrNonNativeFee))
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddrFee, baseDenom).Amount.GTE(fullExpectedOutput.Amount.TruncateInt()))
}
