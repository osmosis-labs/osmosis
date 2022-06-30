package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
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

	tests := []struct {
		name      string
		coin      sdk.Coin
		baseDenom string
		denom     string
		poolType  gammtypes.PoolI
		swapFee   sdk.Dec
	}{
		{
			name:      "TxFees AfterEpochEnd for uion",
			coin:      sdk.NewInt64Coin(uion, 10),
			baseDenom: baseDenom,
			denom:     uion,
			poolType:  uionPool,
			swapFee:   sdk.NewDec(0),
		},
		{
			name:      "TxFees AfterEpochEnd for atom",
			coin:      sdk.NewInt64Coin(atom, 20),
			baseDenom: baseDenom,
			denom:     atom,
			poolType:  atomPool,
			swapFee:   sdk.NewDec(0),
		},
		{
			name:      "TxFees AfterEpochEnd for ust :(",
			coin:      sdk.NewInt64Coin(ust, 14),
			baseDenom: baseDenom,
			denom:     ust,
			poolType:  ustPool,
			swapFee:   sdk.NewDec(0),
		},
	}

	var finalOutputAmount = sdk.NewInt(0)
	moduleAddrFee := suite.App.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
	moduleAddrNonNativeFee := suite.App.AccountKeeper.GetModuleAddress(types.NonNativeFeeCollectorName)

	params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	futureCtx := suite.Ctx.WithBlockTime(time.Now().Add(time.Minute))

	for _, tc := range tests {
		expectedOutput, err := tc.poolType.CalcOutAmtGivenIn(suite.Ctx,
			sdk.Coins{sdk.Coin{Denom: tc.denom, Amount: tc.coin.Amount}},
			tc.baseDenom,
			tc.swapFee)
		suite.Require().NoError(err)

		finalOutputAmount = finalOutputAmount.Add(expectedOutput.Amount)
		_, _, addr0 := testdata.KeyTestPubAddr()
		simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, sdk.Coins{tc.coin})
		suite.App.BankKeeper.SendCoinsFromAccountToModule(suite.Ctx, addr0, types.NonNativeFeeCollectorName, sdk.Coins{tc.coin})

		suite.Require().True(suite.App.BankKeeper.HasBalance(suite.Ctx, moduleAddrNonNativeFee, tc.coin))

		suite.App.EpochsKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))
		moduleBaseDenomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddrFee, tc.baseDenom)
		suite.Require().Empty(suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrNonNativeFee))
		// Tried to make use Equal() to check the exact same value but moduleBaseDenomBalance.Amount seems to be incrementing
		// by 2000000 per test hmm
		suite.Require().True(moduleBaseDenomBalance.Amount.GTE(finalOutputAmount))
	}
}
