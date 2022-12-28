package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

// TestSendDeveloperFeesToDeveloperAccount tests the SendDeveloperFeesToDeveloperAccount function
func (suite *KeeperTestSuite) TestSendDeveloperFeesToDeveloperAccount() {
	cases := []struct {
		description   string
		alterState    func()
		expectedErr   bool
		expectedCoins sdk.Coins
	}{
		{
			description:   "Send with unset developer account",
			alterState:    func() {},
			expectedErr:   true,
			expectedCoins: sdk.NewCoins(),
		},
		{
			description: "Send with set developer account",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 0)
				suite.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(400))),
		},
		{
			description: "Send with set developer account (after multiple trades)",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				// Trade 1
				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 0)
				suite.Require().NoError(err)

				// Trade 2
				err = suite.pseudoExecuteTrade(types.AtomDenomination, sdk.NewInt(2000), 0)
				suite.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(400)), sdk.NewCoin(types.AtomDenomination, sdk.NewInt(400))),
		},
		{
			description: "Send with set developer account (after multiple trades across epochs)",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				// Trade 1
				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 0)
				suite.Require().NoError(err)

				// Trade 2
				err = suite.pseudoExecuteTrade(types.AtomDenomination, sdk.NewInt(2000), 0)
				suite.Require().NoError(err)

				// Trade 3 after year 1
				err = suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 366)
				suite.Require().NoError(err)

				// Trade 4 after year 2
				err = suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 366*2)
				suite.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.AtomDenomination, sdk.NewInt(400)), sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(700))),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			tc.alterState()

			err := suite.App.AppKeepers.ProtoRevKeeper.SendDeveloperFeesToDeveloperAccount(suite.Ctx)
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			developerAccount, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
			if err == nil {
				developerFees := suite.App.AppKeepers.BankKeeper.GetAllBalances(suite.Ctx, developerAccount)
				suite.Require().Equal(tc.expectedCoins, developerFees)
			}
		})
	}
}

// TestUpdateDeveloperFees tests the UpdateDeveloperFees function
func (suite *KeeperTestSuite) TestUpdateDeveloperFees() {
	cases := []struct {
		description string
		denom       string
		profit      sdk.Int
		alterState  func()
		expected    sdk.Coin
	}{
		{
			description: "Update developer fees in year 1",
			denom:       types.OsmosisDenomination,
			profit:      sdk.NewInt(200),
			alterState:  func() {},
			expected:    sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(40)),
		},
		{
			description: "Update developer fees in year 2",
			denom:       types.OsmosisDenomination,
			profit:      sdk.NewInt(200),
			alterState: func() {
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, 366)
			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(20)),
		},
		{
			description: "Update developer fees after year 2",
			denom:       types.OsmosisDenomination,
			profit:      sdk.NewInt(200),
			alterState: func() {
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, 731)

			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
		{
			description: "Update developer fees after year 10",
			denom:       types.OsmosisDenomination,
			profit:      sdk.NewInt(200),
			alterState: func() {
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, 365*10+1)

			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			tc.alterState()

			err := suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, tc.denom, tc.profit)
			suite.Require().NoError(err)

			developerFees, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, tc.denom)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expected, developerFees)
		})
	}
}

// pseudoExecuteTrade is a helper function to execute a trade given denom of profit, profit, and days since genesis
func (suite *KeeperTestSuite) pseudoExecuteTrade(denom string, profit sdk.Int, daysSinceGenesis uint64) error {
	// Initialize the number of days since genesis
	suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, daysSinceGenesis)
	// Mint the profit to the module account (which will be sent to the developer account later)
	suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, profit)))
	// Update the developer fees
	return suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, denom, profit)
}
