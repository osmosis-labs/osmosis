package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

var defaultPooledAssetAmount = int64(500)

func (suite *KeeperTestSuite) preparePool(denom string) (poolID uint64, pool poolmanagertypes.PoolI) {
	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	poolID = suite.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(denom, defaultPooledAssetAmount),
	)
	pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolID)
	suite.Require().NoError(err)
	err = suite.ExecuteUpgradeFeeTokenProposal(denom, poolID)
	suite.Require().NoError(err)
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
		name       string
		coins      sdk.Coins
		baseDenom  string
		denoms     []string
		poolTypes  []poolmanagertypes.PoolI
		swapFee    sdk.Dec
		expectPass bool
	}{
		{
			name:      "One non-osmo fee token (uion): TxFees AfterEpochEnd",
			coins:     sdk.Coins{sdk.NewInt64Coin(uion, 10)},
			baseDenom: baseDenom,
			denoms:    []string{uion},
			poolTypes: []poolmanagertypes.PoolI{uionPool},
			swapFee:   sdk.MustNewDecFromStr("0"),
		},
		{
			name:      "Multiple non-osmo fee token: TxFees AfterEpochEnd",
			coins:     sdk.Coins{sdk.NewInt64Coin(atom, 20), sdk.NewInt64Coin(ust, 30)},
			baseDenom: baseDenom,
			denoms:    []string{atom, ust},
			poolTypes: []poolmanagertypes.PoolI{atomPool, ustPool},
			swapFee:   sdk.MustNewDecFromStr("0"),
		},
	}

	finalOutputAmount := sdk.NewInt(0)

	for _, tc := range tests {
		tc := tc

		suite.Run(tc.name, func() {
			for i, coin := range tc.coins {
				// Get the output amount in osmo denom
				pool, ok := tc.poolTypes[i].(gammtypes.CFMMPoolI)
				suite.Require().True(ok)

				expectedOutput, err := pool.CalcOutAmtGivenIn(suite.Ctx,
					sdk.Coins{sdk.Coin{Denom: tc.denoms[i], Amount: coin.Amount}},
					tc.baseDenom,
					tc.swapFee)
				suite.NoError(err)
				// sanity check for the expectedAmount
				suite.True(coin.Amount.GTE(expectedOutput.Amount))

				finalOutputAmount = finalOutputAmount.Add(expectedOutput.Amount)

				// Deposit some fee amount (non-native-denom) to the fee module account
				_, _, addr0 := testdata.KeyTestPubAddr()
				err = simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, sdk.Coins{coin})
				suite.NoError(err)
				err = suite.App.BankKeeper.SendCoinsFromAccountToModule(suite.Ctx, addr0, types.NonNativeFeeCollectorName, sdk.Coins{coin})
				suite.NoError(err)
			}

			// checks the balance of the non-native denom in module account
			moduleAddrNonNativeFee := suite.App.AccountKeeper.GetModuleAddress(types.NonNativeFeeCollectorName)
			suite.Equal(suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrNonNativeFee), tc.coins)

			// End of epoch, so all the non-osmo fee amount should be swapped to osmo and transfer to fee module account
			params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
			futureCtx := suite.Ctx.WithBlockTime(time.Now().Add(time.Minute))
			err := suite.App.TxFeesKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))
			suite.NoError(err)

			// check the balance of the native-basedenom in module
			moduleAddrFee := suite.App.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
			moduleBaseDenomBalance := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddrFee, tc.baseDenom)

			// non-osmos module account should be empty as all the funds should be transferred to osmo module
			suite.Empty(suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrNonNativeFee))
			// check that the total osmo amount has been transferred to module account
			suite.Equal(moduleBaseDenomBalance.Amount, finalOutputAmount)
		})
	}
}
