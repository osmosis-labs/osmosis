package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	valPref "github.com/osmosis-labs/osmosis/v12/x/validator-preference"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/types"
)

func (suite *KeeperTestSuite) TestSetValidatorSetPreference() {
	suite.SetupTest()

	// setup 3 validators
	valAddrs := suite.SetupMultipleValidators(3)

	tests := []struct {
		name        string
		delegator   sdk.AccAddress
		preferences []types.ValidatorPreference
		creationFee sdk.Coins
		expectPass  bool
	}{
		{
			name:      "creation of new validator set",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(3, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
			},
			creationFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))),
			expectPass:  true,
		},
		{
			name:      "update existing validator with same valAddr and weights",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(3, 1),
				},
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
			},
			creationFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))),
			expectPass:  false,
		},
		{
			name:      "update existing validator with same valAddr but different weights",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(3, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
			},
			creationFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))),
			expectPass:  true,
		},
		{
			name:      "create validator set with unknown validator address",
			delegator: sdk.AccAddress([]byte("addr2---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: "addr1---------------",
					Weight:         sdk.NewDec(1),
				},
			},
			creationFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))),
			expectPass:  false,
		},
		{
			name:      "creation of new validator set with 0 fees",
			delegator: sdk.AccAddress([]byte("addr3---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
			},
			creationFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0))),
			expectPass:  false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {

			bankKeeper := suite.App.BankKeeper

			// fund the account that is trying to delegate
			suite.FundAcc(test.delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)})
			initialBalance := bankKeeper.GetBalance(suite.Ctx, test.delegator, sdk.DefaultBondDenom).Amount

			// set the creation fee
			suite.App.ValidatorPreferenceKeeper.SetParams(suite.Ctx, types.Params{
				ValsetCreationFee: test.creationFee,
			})

			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			// call the create validator set preference
			_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, test.preferences))
			if test.expectPass {
				suite.Require().NoError(err)

				// check if the fee has been used
				balance := bankKeeper.GetBalance(suite.Ctx, test.delegator, sdk.DefaultBondDenom).Amount
				suite.Require().True(balance.LT(initialBalance))
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite *KeeperTestSuite) TestDelegateToValidatorSet() {
	tests := []struct {
		name       string
		delegator  sdk.AccAddress
		coin       sdk.Coin
		expectPass bool
	}{
		{
			name:       "Stake to a valid validator!",
			delegator:  sdk.AccAddress([]byte("addr1---------------")),
			coin:       sdk.NewCoin("stake", sdk.NewInt(10)),
			expectPass: true,
		},
		{
			name:       "User doesnot have enough tokens to stake",
			delegator:  sdk.AccAddress([]byte("addr2---------------")),
			coin:       sdk.NewCoin("stake", sdk.NewInt(200)),
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()

			suite.FundAcc(test.delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)})

			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			// call the create validator set preference
			preferences := suite.PrepareDelegateToValidatorSet()

			// set the creation fee
			suite.App.ValidatorPreferenceKeeper.SetParams(suite.Ctx, types.Params{
				ValsetCreationFee: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))),
			})

			_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
			suite.Require().NoError(err)

			// Fund the delegator address account
			suite.FundAcc(test.delegator, sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(20))})

			// call the create validator set preference
			_, err = msgServer.DelegateToValidatorSet(c, types.NewMsgDelegateToValidatorSet(test.delegator, test.coin))
			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
