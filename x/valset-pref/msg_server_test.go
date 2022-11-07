package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	valPref "github.com/osmosis-labs/osmosis/v12/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v12/x/valset-pref/types"
)

func (suite *KeeperTestSuite) TestSetValidatorSetPreference() {
	suite.SetupTest()

	// setup 3 validators
	valAddrs := suite.SetupMultipleValidators(3)

	tests := []struct {
		name        string
		delegator   sdk.AccAddress
		preferences []types.ValidatorPreference
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
			expectPass: true,
		},
		{
			name:      "update 2 validator weights but leave the 3rd one as is",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(4, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(1, 1),
				},
			},
			expectPass: true,
		},
		{
			name:      "update existing validator with same valAddr and weights",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(4, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(1, 1),
				},
			},
			expectPass: false,
		},
		{
			name:      "update existing validator with same valAddr but different weights",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(1, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(7, 1),
				},
			},
			expectPass: true,
		},
		{
			name:      "create validator set with unknown validator address",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			preferences: []types.ValidatorPreference{
				{
					ValOperAddress: "addr1---------------",
					Weight:         sdk.NewDec(1),
				},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			// call the create validator set preference
			_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, test.preferences))
			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite *KeeperTestSuite) TestDelegateToValidatorSet() {
	suite.SetupTest()

	// prepare validators to delegate to
	preferences := suite.PrepareDelegateToValidatorSet()

	tests := []struct {
		name           string
		delegator      sdk.AccAddress
		coin           sdk.Coin
		expectedShares []sdk.Dec
		expectPass     bool
		valSetExists   bool
	}{
		{
			name:           "Delegate to valid validators!",
			delegator:      sdk.AccAddress([]byte("addr1---------------")),
			coin:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),                      // amount to delegate
			expectedShares: []sdk.Dec{sdk.NewDec(5_000_000), sdk.NewDec(3_000_000), sdk.NewDec(2_000_000)}, // expected shares after delegation
			expectPass:     true,
		},
		{
			name:           "Delegate more tokens to existing validator-set",
			delegator:      sdk.AccAddress([]byte("addr1---------------")),
			coin:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),                       // amount to delegate
			expectedShares: []sdk.Dec{sdk.NewDec(10_000_000), sdk.NewDec(6_000_000), sdk.NewDec(4_000_000)}, // expected shares after delegation
			expectPass:     true,
			valSetExists:   true,
		},
		{
			name:           "Delegate Decimal Amounts",
			delegator:      sdk.AccAddress([]byte("addr2---------------")),
			coin:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(25_000_000)),                       // amount to delegate
			expectedShares: []sdk.Dec{sdk.NewDec(12_500_000), sdk.NewDec(7_500_000), sdk.NewDec(5_000_000)}, // expected shares after delegation
			expectPass:     true,
		},
		{
			name:       "User doesnot have enough tokens to stake",
			delegator:  sdk.AccAddress([]byte("addr3---------------")),
			coin:       sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000_000)), // amount to delegate
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.FundAcc(test.delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)}) // 100 osmo

			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			if !test.valSetExists {
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				suite.Require().NoError(err)
			}

			// call the create validator set preference
			_, err := msgServer.DelegateToValidatorSet(c, types.NewMsgDelegateToValidatorSet(test.delegator, test.coin))
			if test.expectPass {
				suite.Require().NoError(err)

				// check if the expectedShares matches after delegation
				for i, val := range preferences {
					valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
					suite.Require().NoError(err)

					//gurantees that the delegator exist because we check it in DelegateToValidatorSet
					del, _ := suite.App.StakingKeeper.GetDelegation(suite.Ctx, test.delegator, valAddr)
					suite.Require().Equal(del.Shares, test.expectedShares[i])
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}
