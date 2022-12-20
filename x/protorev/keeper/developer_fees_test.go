package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

// TestSendDeveloperFeesToDeveloperAccoutn tests the SendDeveloperFeesToDeveloperAccount function
func (suite *KeeperTestSuite) TestSendDeveloperFeesToDeveloperAccount() {
	cases := []struct {
		description   string
		malleate      func()
		expectedErr   bool
		expectedCoins sdk.Coins
	}{
		{
			description:   "Send with unset developer account",
			malleate:      func() {},
			expectedErr:   true,
			expectedCoins: sdk.NewCoins(),
		},
		{
			description: "Send with set developer account",
			malleate: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000))))
				err := suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))
				suite.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(200))),
		},
		{
			description: "Send with set developer account (after multiple trades)",
			malleate: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				// Trade 1
				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000))))
				err := suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))
				suite.Require().NoError(err)

				// Trade 2
				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.AtomDenomination, sdk.NewInt(1000))))
				err = suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.AtomDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))
				suite.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(200)), sdk.NewCoin(types.AtomDenomination, sdk.NewInt(200))),
		},
		{
			description: "Send with set developer account (after multiple trades across epochs)",
			malleate: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.AppKeepers.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				// Trade 1
				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000))))
				err := suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))
				suite.Require().NoError(err)

				// Trade 2
				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.AtomDenomination, sdk.NewInt(1000))))
				err = suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.AtomDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))
				suite.Require().NoError(err)

				// Trade 3 after year 1
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceGenesis(suite.Ctx, 366)
				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000))))
				err = suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))

				// Trade 4 after year 2
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceGenesis(suite.Ctx, 366*2)
				suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000))))
				err = suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)), sdk.NewInt(2000))
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.AtomDenomination, sdk.NewInt(200)), sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(350))),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			tc.malleate()

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
		description    string
		inputCoin      sdk.Coin
		tokenOutAmount sdk.Int
		malleate       func()
		expected       sdk.Coin
	}{
		{
			description:    "Update developer fees in year 1",
			inputCoin:      sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)),
			tokenOutAmount: sdk.NewInt(200),
			malleate:       func() {},
			expected:       sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(20)),
		},
		{
			description:    "Update developer fees in year 2",
			inputCoin:      sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)),
			tokenOutAmount: sdk.NewInt(200),
			malleate: func() {
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceGenesis(suite.Ctx, 366)
			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
		{
			description:    "Update developer fees after year 2",
			inputCoin:      sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)),
			tokenOutAmount: sdk.NewInt(200),
			malleate: func() {
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceGenesis(suite.Ctx, 731)

			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(5)),
		},
		{
			description:    "Update developer fees after year 10",
			inputCoin:      sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)),
			tokenOutAmount: sdk.NewInt(200),
			malleate: func() {
				suite.App.AppKeepers.ProtoRevKeeper.SetDaysSinceGenesis(suite.Ctx, 365*10+1)

			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(5)),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			tc.malleate()

			err := suite.App.AppKeepers.ProtoRevKeeper.UpdateDeveloperFees(suite.Ctx, tc.inputCoin, tc.tokenOutAmount)
			suite.Require().NoError(err)

			developerFees, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperFees(suite.Ctx, tc.inputCoin.Denom)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expected, developerFees)
		})
	}
}
