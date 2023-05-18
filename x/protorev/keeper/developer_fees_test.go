package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// TestSendDeveloperFeesToDeveloperAccount tests the SendDeveloperFeesToDeveloperAccount function
func (s *KeeperTestSuite) TestSendDeveloperFeesToDeveloperAccount() {
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
				s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

				err := s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 0)
				s.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(400))),
		},
		{
			description: "Send with set developer account (after multiple trades)",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

				// Trade 1
				err := s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 0)
				s.Require().NoError(err)

				// Trade 2
				err = s.pseudoExecuteTrade("Atom", sdk.NewInt(2000), 0)
				s.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(400)), sdk.NewCoin("Atom", sdk.NewInt(400))),
		},
		{
			description: "Send with set developer account (after multiple trades across epochs)",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

				// Trade 1
				err := s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 0)
				s.Require().NoError(err)

				// Trade 2
				err = s.pseudoExecuteTrade("Atom", sdk.NewInt(2000), 0)
				s.Require().NoError(err)

				// Trade 3 after year 1
				err = s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 366)
				s.Require().NoError(err)

				// Trade 4 after year 2
				err = s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(2000), 366*2)
				s.Require().NoError(err)
			},
			expectedErr:   false,
			expectedCoins: sdk.NewCoins(sdk.NewCoin("Atom", sdk.NewInt(400)), sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(700))),
		},
	}

	for _, tc := range cases {
		s.Run(tc.description, func() {
			s.SetupTest()
			tc.alterState()

			err := s.App.ProtoRevKeeper.SendDeveloperFeesToDeveloperAccount(s.Ctx)
			if tc.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			developerAccount, err := s.App.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
			if !tc.expectedErr {
				developerFees := s.App.AppKeepers.BankKeeper.GetAllBalances(s.Ctx, developerAccount)
				s.Require().Equal(tc.expectedCoins, developerFees)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestUpdateDeveloperFees tests the UpdateDeveloperFees function
func (s *KeeperTestSuite) TestUpdateDeveloperFees() {
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
				s.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(s.Ctx, 366)
			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(20)),
		},
		{
			description: "Update developer fees after year 2",
			denom:       types.OsmosisDenomination,
			profit:      sdk.NewInt(200),
			alterState: func() {
				s.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(s.Ctx, 731)
			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
		{
			description: "Update developer fees after year 10",
			denom:       types.OsmosisDenomination,
			profit:      sdk.NewInt(200),
			alterState: func() {
				s.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(s.Ctx, 365*10+1)
			},
			expected: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
	}

	for _, tc := range cases {
		s.Run(tc.description, func() {
			s.SetupTest()
			tc.alterState()

			err := s.App.ProtoRevKeeper.UpdateDeveloperFees(s.Ctx, tc.denom, tc.profit)
			s.Require().NoError(err)

			developerFees, err := s.App.ProtoRevKeeper.GetDeveloperFees(s.Ctx, tc.denom)
			s.Require().NoError(err)
			s.Require().Equal(tc.expected, developerFees)
		})
	}
}

// pseudoExecuteTrade is a helper function to execute a trade given denom of profit, profit, and days since genesis
func (s *KeeperTestSuite) pseudoExecuteTrade(denom string, profit sdk.Int, daysSinceGenesis uint64) error {
	// Initialize the number of days since genesis
	s.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(s.Ctx, daysSinceGenesis)
	// Mint the profit to the module account (which will be sent to the developer account later)
	err := s.App.AppKeepers.BankKeeper.MintCoins(s.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, profit)))
	if err != nil {
		return err
	}
	// Update the developer fees
	return s.App.ProtoRevKeeper.UpdateDeveloperFees(s.Ctx, denom, profit)
}
