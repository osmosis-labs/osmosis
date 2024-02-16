package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting"
	"github.com/osmosis-labs/osmosis/v23/x/protorev/types"
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
		expectedModuleBal sdk.Coin
		expectedBurnBal   sdk.Coin
	}{
		{
			description:       "Send with unset developer account",
			alterState:        func() {},
			denom:             types.OsmosisDenomination,
			expectedErr:       true,
			expectedDevProfit: sdk.Coin{},
			expectedModuleBal: sdk.Coin{},
			expectedBurnBal:   sdk.Coin{},
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
			expectedModuleBal: sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(0)),
			expectedBurnBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(800)),
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
			expectedModuleBal: sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(0)),
			expectedBurnBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(900)),
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
			expectedModuleBal: sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(0)),
			expectedBurnBal:   sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(950)),
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
			expectedModuleBal: sdk.NewCoin(usdcDenom, osmomath.NewInt(800)),
			expectedBurnBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(0)),
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
			expectedModuleBal: sdk.NewCoin(usdcDenom, osmomath.NewInt(900)),
			expectedBurnBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(0)),
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
			expectedModuleBal: sdk.NewCoin(usdcDenom, osmomath.NewInt(950)),
			expectedBurnBal:   sdk.NewCoin(usdcDenom, osmomath.NewInt(0)),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			tc.alterState()

			err := suite.App.ProtoRevKeeper.DistributeProfit(suite.Ctx, sdk.NewCoin(tc.denom, arbProfit))
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			developerAccount, err := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
			if !tc.expectedErr {
				// Validate the developer account balance.
				developerFee := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, developerAccount, tc.denom)
				suite.Require().Equal(tc.expectedDevProfit, developerFee)

				// Validate the module account balance.
				moduleAccount := suite.App.AppKeepers.AccountKeeper.GetModuleAddress(types.ModuleName)
				moduleBal := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, moduleAccount, tc.denom)
				suite.Require().Equal(tc.expectedModuleBal, moduleBal)

				// Validate the burn balance.
				burnBal := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, types.DefaultNullAddress, tc.denom)
				suite.Require().Equal(tc.expectedBurnBal, burnBal)
			} else {
				suite.Require().Error(err)
			}
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
