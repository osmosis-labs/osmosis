package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	"github.com/osmosis-labs/osmosis/v16/x/protorev/types"
)

// TestSendDeveloperFee tests the SendDeveloperFee function
func (s *KeeperTestSuite) TestSendDeveloperFee() {
	cases := []struct {
		description       string
		alterState        func()
		expectedErr       bool
		expectedDevProfit sdk.Coin
	}{
		{
			description:       "Send with unset developer account",
			alterState:        func() {},
			expectedErr:       true,
			expectedDevProfit: sdk.Coin{},
		},
		{
			description: "Send with set developer account in first phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

				err := s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(1000), 100)
				s.Require().NoError(err)
			},
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(20)),
		},
		{
			description: "Send with set developer account in second phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

				err := s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(1000), 500)
				s.Require().NoError(err)
			},
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
		{
			description: "Send with set developer account in third (final) phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				s.App.ProtoRevKeeper.SetDeveloperAccount(s.Ctx, account)

				err := s.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(1000), 1000)
				s.Require().NoError(err)
			},
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(5)),
		},
	}

	for _, tc := range cases {
		s.Run(tc.description, func() {
			s.SetupTest()
			tc.alterState()

			err := s.App.ProtoRevKeeper.SendDeveloperFee(s.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)))
			if tc.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			developerAccount, err := s.App.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
			if !tc.expectedErr {
				developerFee := s.App.AppKeepers.BankKeeper.GetBalance(s.Ctx, developerAccount, types.OsmosisDenomination)
				s.Require().Equal(tc.expectedDevProfit, developerFee)
			} else {
				s.Require().Error(err)
			}
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

	return nil
}
