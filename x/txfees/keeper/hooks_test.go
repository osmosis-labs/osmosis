package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

func (suite *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	uion := "uion"
	atom := "atom"

	baseDenom := sdk.DefaultBondDenom

	tests := []struct {
		name       string
		coins      sdk.Coins
		expectPass bool
	}{
		{
			name:       "DYM is burned",
			coins:      sdk.Coins{sdk.NewInt64Coin(baseDenom, 100000)},
			expectPass: true,
		},
		{
			name:       "One non-dym fee token (uion)",
			coins:      sdk.Coins{sdk.NewInt64Coin(uion, 1000)},
			expectPass: true,
		},
		{
			name:       "Multiple non-dym fee token",
			coins:      sdk.Coins{sdk.NewInt64Coin(baseDenom, 2000), sdk.NewInt64Coin(uion, 30000)},
			expectPass: true,
		},
		{
			name:       "unknown fee token",
			coins:      sdk.Coins{sdk.NewInt64Coin(atom, 2000)},
			expectPass: false,
		},
	}

	for _, tc := range tests {
		suite.SetupTest(false)

		// create pools for three separate fee tokens
		suite.PrepareBalancerPoolWithCoins(sdk.NewCoin(baseDenom, sdk.NewInt(1000000000000)), sdk.NewCoin(uion, sdk.NewInt(5000)))

		moduleAddrFee := suite.App.AccountKeeper.GetModuleAddress(types.ModuleName)
		err := bankutil.FundModuleAccount(suite.App.BankKeeper, suite.Ctx, types.ModuleName, tc.coins)
		suite.Require().NoError(err)
		balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrFee)
		suite.Assert().Equal(balances, tc.coins, tc.name)

		totalSupplyBefore := suite.App.BankKeeper.GetSupply(suite.Ctx, baseDenom).Amount

		// End of epoch, so all the non-osmo fee amount should be swapped to osmo and burned
		futureCtx := suite.Ctx.WithBlockTime(time.Now().Add(time.Minute))
		suite.App.TxFeesKeeper.AfterEpochEnd(futureCtx, types.EpochIdentifier, int64(1))

		// check the balance of the native-basedenom in module
		balances = suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddrFee)
		totalSupplyAfter := suite.App.BankKeeper.GetSupply(suite.Ctx, baseDenom).Amount
		if tc.expectPass {
			//Check for DYM burn
			suite.Assert().True(balances.IsZero(), tc.name)
			suite.Require().True(totalSupplyAfter.LT(totalSupplyBefore), tc.name)
		} else {
			suite.Assert().False(balances.IsZero(), tc.name)
			suite.Require().True(totalSupplyAfter.Equal(totalSupplyBefore), tc.name)
		}
	}
}
