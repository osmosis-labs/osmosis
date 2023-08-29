package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v19/app/apptesting"
	"github.com/osmosis-labs/osmosis/v19/x/protorev/types"
)

// TestSendDeveloperFee tests the SendDeveloperFee function
func (suite *KeeperTestSuite) TestSendDeveloperFee() {
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
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(1000), 100)
				suite.Require().NoError(err)
			},
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(20)),
		},
		{
			description: "Send with set developer account in second phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(1000), 500)
				suite.Require().NoError(err)
			},
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10)),
		},
		{
			description: "Send with set developer account in third (final) phase",
			alterState: func() {
				account := apptesting.CreateRandomAccounts(1)[0]
				suite.App.ProtoRevKeeper.SetDeveloperAccount(suite.Ctx, account)

				err := suite.pseudoExecuteTrade(types.OsmosisDenomination, sdk.NewInt(1000), 1000)
				suite.Require().NoError(err)
			},
			expectedErr:       false,
			expectedDevProfit: sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(5)),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			tc.alterState()

			err := suite.App.ProtoRevKeeper.SendDeveloperFee(suite.Ctx, sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(100)))
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			developerAccount, err := suite.App.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
			if !tc.expectedErr {
				developerFee := suite.App.AppKeepers.BankKeeper.GetBalance(suite.Ctx, developerAccount, types.OsmosisDenomination)
				suite.Require().Equal(tc.expectedDevProfit, developerFee)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// pseudoExecuteTrade is a helper function to execute a trade given denom of profit, profit, and days since genesis
func (suite *KeeperTestSuite) pseudoExecuteTrade(denom string, profit sdk.Int, daysSinceGenesis uint64) error {
	// Initialize the number of days since genesis
	suite.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, daysSinceGenesis)
	// Mint the profit to the module account (which will be sent to the developer account later)
	err := suite.App.AppKeepers.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, profit)))
	if err != nil {
		return err
	}

	return nil
}
