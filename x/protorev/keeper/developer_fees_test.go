package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

var (
	usdcDenom = "usdc"
	arbProfit = osmomath.NewInt(1000)
)

func (suite *KeeperTestSuite) TestDistributeProfit() {
	cases := []struct {
		description       string
		alterState        func()
		denom             string
		expectedErr       bool
		expectedDevProfit sdk.Coin
		expectedBurnBal   sdk.Coin
		expectedCommBal   sdk.Coin
	}{
		{
			description:       "Send with unset developer account",
			alterState:        func() {},
			denom:             types.OsmosisDenomination,
			expectedErr:       true,
			expectedDevProfit: sdk.Coin{},
			expectedBurnBal:   sdk.Coin{},
			expectedCommBal:   sdk.Coin{},
		},
		{
			description: "Send with set developer account in first phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, arbProfit, 100)
				suite.Require().NoError(err)
			},
			denom:             types.OsmosisDenomination,
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(200)),
			expectedBurnBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(800)),
			expectedCommBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(0)),
		},
		{
			description: "Send with set developer account in second phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, arbProfit, 500)
				suite.Require().NoError(err)
			},
			denom:             types.OsmosisDenomination,
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(100)),
			expectedBurnBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(900)),
			expectedCommBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(0)),
		},
		{
			description: "Send with set developer account in third (final) phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, arbProfit, 1000)
				suite.Require().NoError(err)
			},
			denom:             types.OsmosisDenomination,
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(50)),
			expectedBurnBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(950)),
			expectedCommBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(0)),
		},
		{
			description: "Send with set developer account in first phase with non-osmo profit",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(usdcDenom, arbProfit, 100)
				suite.Require().NoError(err)
			},
			denom:             usdcDenom,
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(usdcDenom, osmomath.NewInt(200)),
			expectedBurnBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(0)),
			expectedCommBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(800)),
		},
		{
			description: "Send with set developer account in second phase with non-osmo profit",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(usdcDenom, arbProfit, 500)
				suite.Require().NoError(err)
			},
			denom:             usdcDenom,
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(usdcDenom, osmomath.NewInt(100)),
			expectedBurnBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(0)),
			expectedCommBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(900)),
		},
		{
			description: "Send with set developer account in third (final) phase with non-osmo profit",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(usdcDenom, arbProfit, 1000)
				suite.Require().NoError(err)
			},
			denom:             usdcDenom,
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(usdcDenom, osmomath.NewInt(50)),
			expectedBurnBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(0)),
			expectedCommBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(950)),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupNoPools()

			commAccount := suite.App.AppKeepers.AccountKeeper.GetModuleAddress(distributiontypes.ModuleName)
			commBalanceBefore := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, commAccount, tc.denom)

			tc.alterState()

			err := suite.App.ProtoRevKeeper.DistributeProfit(suite.Ctx, sdk.NewCoins(sdk.NewCoin(tc.denom, arbProfit)))
			if tc.expectedErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// Validate the developer account balance.
			developerAccount, err := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
			suite.Require().NoError(err)
			developerFee := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, developerAccount, tc.denom)
			suite.Require().True(tc.expectedDevProfit.Equal(developerFee))

			// Validate the module community pool balance.
			commBalanceAfter := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, commAccount, tc.denom)
			diff := commBalanceAfter.Sub(commBalanceBefore)
			suite.Require().True(tc.expectedCommBal.Equal(diff))

			// Validate the burn balance.
			burnBal := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, types.DefaultNullAddress, tc.denom)
			suite.Require().True(tc.expectedBurnBal.Equal(burnBal))
		})
	}
}

// pseudoExecuteTrade is a helper function to execute a trade given denom of profit, profit, and days since genesis
func (suite *KeeperTestSuite) pseudoExecuteTrade(denom string, profit osmomath.Int, daysSinceGenesis uint64) error {
	// Initialize the number of days since genesis
	suite.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, daysSinceGenesis)
	// Mint the profit to the module account (which will be sent to the developer account later)
	err := suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, profit)))
	if err != nil {
		return err
	}

	return nil
}
