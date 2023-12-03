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

//TODO: pool hooks
/*

func (suite *KeeperTestSuite) TestUpgradeFeeTokenProposals() {
	suite.SetupTest(false)

	uionPoolId := suite.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	uionPoolId2 := suite.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	// Make pool with fee token but no OSMO and make sure governance proposal fails
	noBasePoolId := suite.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("uion", 500),
		sdk.NewInt64Coin("foo", 500),
	)

	// Create correct pool and governance proposal
	fooPoolId := suite.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("foo", 1000),
	)

	tests := []struct {
		name       string
		feeToken   string
		poolId     uint64
		expectPass bool
	}{
		{
			name:       "uion pool",
			feeToken:   "uion",
			poolId:     uionPoolId,
			expectPass: true,
		},
		{
			name:       "try with basedenom",
			feeToken:   sdk.DefaultBondDenom,
			poolId:     uionPoolId,
			expectPass: false,
		},
		{
			name:       "proposal with non-existent pool",
			feeToken:   "foo",
			poolId:     100000000000,
			expectPass: false,
		},
		{
			name:       "proposal with wrong pool for fee token",
			feeToken:   "foo",
			poolId:     uionPoolId,
			expectPass: false,
		},
		{
			name:       "proposal with pool with no base denom",
			feeToken:   "foo",
			poolId:     noBasePoolId,
			expectPass: false,
		},
		{
			name:       "proposal to add foo correctly",
			feeToken:   "foo",
			poolId:     fooPoolId,
			expectPass: true,
		},
		{
			name:       "proposal to replace pool for fee token",
			feeToken:   "uion",
			poolId:     uionPoolId2,
			expectPass: true,
		},
		{
			name:       "proposal to replace uion as fee denom",
			feeToken:   "uion",
			poolId:     0,
			expectPass: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			feeTokensBefore := suite.App.TxFeesKeeper.GetFeeTokens(suite.Ctx)

			// Add a new whitelisted fee token via a governance proposal
			err := suite.ExecuteUpgradeFeeTokenProposal(tc.feeToken, tc.poolId)

			feeTokensAfter := suite.App.TxFeesKeeper.GetFeeTokens(suite.Ctx)

			if tc.expectPass {
				// Make sure no error during setting of proposal
				suite.Require().NoError(err, "test: %s", tc.name)

				// For a proposal that adds a feetoken
				if tc.poolId != 0 {
					// Make sure the length of fee tokens is >= before
					suite.Require().GreaterOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
					// Ensure that the fee token is convertable to base token
					_, err := suite.App.TxFeesKeeper.ConvertToBaseToken(suite.Ctx, sdk.NewInt64Coin(tc.feeToken, 10))
					suite.Require().NoError(err, "test: %s", tc.name)
					// make sure the queried poolId is the same as expected
					queriedPoolId, err := suite.queryClient.DenomPoolId(suite.Ctx.Context(),
						&types.QueryDenomPoolIdRequest{
							Denom: tc.feeToken,
						},
					)
					suite.Require().NoError(err, "test: %s", tc.name)
					suite.Require().Equal(tc.poolId, queriedPoolId.GetPoolID(), "test: %s", tc.name)
				} else {
					// if this proposal deleted a fee token
					// ensure that the length of fee tokens is <= to before
					suite.Require().LessOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
					// Ensure that the fee token is not convertable to base token
					_, err := suite.App.TxFeesKeeper.ConvertToBaseToken(suite.Ctx, sdk.NewInt64Coin(tc.feeToken, 10))
					suite.Require().Error(err, "test: %s", tc.name)
					// make sure the queried poolId errors
					_, err = suite.queryClient.DenomPoolId(suite.Ctx.Context(),
						&types.QueryDenomPoolIdRequest{
							Denom: tc.feeToken,
						},
					)
					suite.Require().Error(err, "test: %s", tc.name)
				}
			} else {
				// Make sure errors during setting of proposal
				suite.Require().Error(err, "test: %s", tc.name)
				// fee tokens should be the same
				suite.Require().Equal(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
			}
		})
	}
}
*/
