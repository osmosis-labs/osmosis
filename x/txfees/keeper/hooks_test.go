package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v11/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v11/x/txfees/types"
)

var defaultPooledAssetAmount = int64(500)

func (suite *KeeperTestSuite) preparePool(denom string) (poolID uint64, pool gammtypes.PoolI) {
	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	poolID = suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(denom, defaultPooledAssetAmount),
	)
	pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolID)
	suite.Require().NoError(err)
	suite.ExecuteUpgradeFeeTokenProposal(denom, poolID)
	return poolID, pool
}

func (suite *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	suite.SetupTest(false)
	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)

	// create pools for three separate fee tokens
	uion := "uion"
	_, uionPool := suite.preparePool(uion)
	atom := "atom"
	_, atomPool := suite.preparePool(atom)
	ust := "ust"
	_, ustPool := suite.preparePool(ust)

	// todo make this section onwards table driven
	coins := sdk.NewCoins(sdk.NewInt64Coin(uion, 10),
		sdk.NewInt64Coin(atom, 20),
		sdk.NewInt64Coin(ust, 14))

	swapFee := sdk.NewDec(0)

	expectedOutput1, err := uionPool.CalcOutAmtGivenIn(suite.Ctx,
		sdk.Coins{sdk.Coin{Denom: uion, Amount: coins.AmountOf(uion)}},
		baseDenom,
		swapFee)
	suite.Require().NoError(err)
	expectedOutput2, err := atomPool.CalcOutAmtGivenIn(suite.Ctx,
		sdk.Coins{sdk.Coin{Denom: atom, Amount: coins.AmountOf(atom)}},
		baseDenom,
		swapFee)
	suite.Require().NoError(err)
	expectedOutput3, err := ustPool.CalcOutAmtGivenIn(suite.Ctx,
		sdk.Coins{sdk.Coin{Denom: ust, Amount: coins.AmountOf(ust)}},
		baseDenom,
		swapFee)
	suite.Require().NoError(err)

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

	moduleBaseDenomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddrFee, baseDenom)
	suite.Require().Empty(suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrNonNativeFee))
	suite.Require().True(moduleBaseDenomBalance.Amount.GTE(fullExpectedOutput.Amount))
}
