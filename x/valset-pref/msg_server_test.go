package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	valPref "github.com/osmosis-labs/osmosis/v14/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
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

	// prepare existing delegations validators
	valAddrs := suite.SetupMultipleValidators(3)

	// prepare validators to delegate to valset
	preferences := suite.PrepareDelegateToValidatorSet()

	amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)} // 100 osmo

	tests := []struct {
		name                   string
		delegator              sdk.AccAddress
		coin                   sdk.Coin  // amount to delegate
		expectedShares         []sdk.Dec // expected shares after delegation
		setExistingDelegations bool
		setValSet              bool
		expectPass             bool
	}{
		{
			name:           "Delegate to valid validators",
			delegator:      sdk.AccAddress([]byte("addr1---------------")),
			coin:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),
			expectedShares: []sdk.Dec{sdk.NewDec(2_000_000), sdk.NewDec(3_300_000), sdk.NewDec(1_200_000), sdk.NewDec(3_500_000)},
			expectPass:     true,
		},
		{
			name:           "Delegate more tokens to existing validator-set",
			delegator:      sdk.AccAddress([]byte("addr1---------------")),
			coin:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),
			expectedShares: []sdk.Dec{sdk.NewDec(4_000_000), sdk.NewDec(6_600_000), sdk.NewDec(2_400_000), sdk.NewDec(7_000_000)},
			expectPass:     true,
		},
		{
			name:       "User does not have enough tokens to stake",
			delegator:  sdk.AccAddress([]byte("addr2---------------")),
			coin:       sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000_000)),
			setValSet:  true,
			expectPass: false,
		},
		{
			name:                   "Delegate to existing staking position (non valSet) ",
			delegator:              sdk.AccAddress([]byte("addr3---------------")),
			coin:                   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),
			expectedShares:         []sdk.Dec{sdk.NewDec(13_333_333), sdk.NewDec(13_333_333), sdk.NewDec(13_333_333)},
			setExistingDelegations: true,
			expectPass:             true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			suite.FundAcc(test.delegator, amountToFund)

			// if validatorSetExist no need to refund and setValSet again
			if test.setValSet {
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				suite.Require().NoError(err)
			}

			if test.setExistingDelegations {
				err := suite.PrepareExistingDelegations(suite.Ctx, valAddrs, test.delegator, test.coin.Amount)
				suite.Require().NoError(err)
			}

			// call the create validator set preference
			_, err := msgServer.DelegateToValidatorSet(c, types.NewMsgDelegateToValidatorSet(test.delegator, test.coin))
			if test.expectPass {
				suite.Require().NoError(err)

				// check if the user balance decreased
				balance := suite.App.BankKeeper.GetBalance(suite.Ctx, test.delegator, sdk.DefaultBondDenom)
				expectedBalance := amountToFund[0].Amount.Sub(test.coin.Amount)
				// valset has not been set so use the (expectedBalance = account balance)
				if !test.setValSet {
					expectedBalance = balance.Amount
				}

				suite.Require().Equal(expectedBalance, balance.Amount)

				if test.setValSet {
					// check if the expectedShares matches after delegation
					for i, val := range preferences {
						valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
						suite.Require().NoError(err)

						// guarantees that the delegator exists because we check it in DelegateToValidatorSet
						del, _ := suite.App.StakingKeeper.GetDelegation(suite.Ctx, test.delegator, valAddr)
						suite.Require().Equal(del.Shares, test.expectedShares[i])
					}
				}

				if test.setExistingDelegations {
					for i, val := range valAddrs {
						valAddr, err := sdk.ValAddressFromBech32(val)
						suite.Require().NoError(err)

						// guarantees that the delegator exists because we check it in DelegateToValidatorSet
						del, _ := suite.App.StakingKeeper.GetDelegation(suite.Ctx, test.delegator, valAddr)
						suite.Require().Equal(del.Shares, test.expectedShares[i])
					}
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUnDelegateFromValidatorSet() {
	suite.SetupTest()

	// prepare existing delegations validators
	valAddrs := suite.SetupMultipleValidators(3)

	// creates a validator preference list to delegate to
	preferences := suite.PrepareDelegateToValidatorSet()

	amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)} // 100 osmo

	tests := []struct {
		name                   string
		delegator              sdk.AccAddress
		coinToStake            sdk.Coin
		coinToUnStake          sdk.Coin
		expectedShares         []sdk.Dec // expected shares after undelegation
		setValSet              bool
		setExistingDelegations bool
		expectPass             bool
	}{
		{
			name:           "Unstake half from the ValSet",
			delegator:      sdk.AccAddress([]byte("addr1---------------")),
			coinToStake:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)), // delegate 20osmo
			coinToUnStake:  sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)), // undelegate 10osmo
			expectedShares: []sdk.Dec{sdk.NewDec(2_000_000), sdk.NewDec(3_300_000), sdk.NewDec(1_200_000), sdk.NewDec(3_500_000)},
			expectPass:     true,
		},
		{
			name:           "Unstake x amount from ValSet",
			delegator:      sdk.AccAddress([]byte("addr2---------------")),
			coinToStake:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),                                           // delegate 20osmo
			coinToUnStake:  sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(15_000_000)),                                           // undelegate 15osmo
			expectedShares: []sdk.Dec{sdk.NewDec(1_000_000), sdk.NewDec(1_650_000), sdk.NewDec(600_000), sdk.NewDec(1_750_000)}, // validatorDelegatedShares - (weight * coinToUnstake)
			expectPass:     true,
		},
		{
			name:          "Unstake everything",
			delegator:     sdk.AccAddress([]byte("addr3---------------")),
			coinToStake:   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			coinToUnStake: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			setValSet:     true,
			expectPass:    true,
		},
		{
			name:          "Unstake more amount than the staked amount",
			delegator:     sdk.AccAddress([]byte("addr4---------------")),
			coinToStake:   sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			coinToUnStake: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(40_000_000)),
			setValSet:     true,
			expectPass:    false,
		},
		{
			name:                   "UnDelegate from existing staking position (non valSet) ",
			delegator:              sdk.AccAddress([]byte("addr5---------------")),
			coinToStake:            sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),
			coinToUnStake:          sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000)),
			expectedShares:         []sdk.Dec{sdk.NewDec(1_000_000), sdk.NewDec(1_660_000), sdk.NewDec(600_000), sdk.NewDec(1_740_000)}, // validatorDelegatedShares - (weight * coinToUnstake)
			setExistingDelegations: true,
			expectPass:             true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.FundAcc(test.delegator, amountToFund) // 100 osmo

			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			if test.setValSet {
				// SetValidatorSetPreference sets a new list of val-set
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				suite.Require().NoError(err)

				// DelegateToValidatorSet delegate to existing val-set
				_, err = msgServer.DelegateToValidatorSet(c, types.NewMsgDelegateToValidatorSet(test.delegator, test.coinToStake))
				suite.Require().NoError(err)
			}

			if test.setExistingDelegations {
				err := suite.PrepareExistingDelegations(suite.Ctx, valAddrs, test.delegator, test.coinToStake.Amount)
				suite.Require().NoError(err)
			}

			_, err := msgServer.UndelegateFromValidatorSet(c, types.NewMsgUndelegateFromValidatorSet(test.delegator, test.coinToUnStake))
			if test.expectPass {
				suite.Require().NoError(err)

				// check if the expectedShares matches after undelegation
				for i, val := range preferences {
					valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
					suite.Require().NoError(err)

					// guarantees that the delegator exists because we check it in UnDelegateToValidatorSet
					del, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, test.delegator, valAddr)
					if found {
						suite.Require().Equal(del.GetShares(), test.expectedShares[i])
					}
				}

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRedelegateValidatorSet() {
	suite.SetupTest()

	// setup 9 validators
	valAddrs := suite.SetupMultipleValidators(9)

	tests := []struct {
		name            string
		delegator       sdk.AccAddress
		newPreferences  []types.ValidatorPreference
		coinToStake     sdk.Coin
		expectedShares  []sdk.Dec // expected shares after redelegation
		delegationExist bool
		expectPass      bool
	}{
		{
			name:      "redelegate to a new set of validators",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			newPreferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(6, 1),
				},
			},
			coinToStake:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			expectedShares: []sdk.Dec{sdk.NewDec(4_000_000), sdk.NewDec(4_000_000), sdk.NewDec(12_000_000)},
			expectPass:     true,
		},
		{
			name:      "redelegate to the same set of validators with different weights, same delegator",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			newPreferences: []types.ValidatorPreference{
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
			coinToStake:     sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			expectedShares:  []sdk.Dec{sdk.NewDec(10_000_000), sdk.NewDec(6_000_000), sdk.NewDec(4_000_000)},
			expectPass:      false,
			delegationExist: true,
		},
		{
			name:      "redelegate to the different set of validators different weights, same delegator",
			delegator: sdk.AccAddress([]byte("addr1---------------")),
			newPreferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[3],
					Weight:         sdk.NewDecWithPrec(5, 1),
				},
				{
					ValOperAddress: valAddrs[4],
					Weight:         sdk.NewDecWithPrec(3, 1),
				},
				{
					ValOperAddress: valAddrs[5],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
			},
			coinToStake:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			expectedShares: []sdk.Dec{sdk.NewDec(10_000_000), sdk.NewDec(6_000_000), sdk.NewDec(4_000_000)},
			expectPass:     true,
		},
		{
			name:      "redelegate to new set, but one validator from old set with different delegator",
			delegator: sdk.AccAddress([]byte("addr2---------------")),
			newPreferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[3],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[4],
					Weight:         sdk.NewDecWithPrec(6, 1),
				},
			},
			coinToStake:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			expectedShares: []sdk.Dec{sdk.NewDec(4_000_000), sdk.NewDec(4_000_000), sdk.NewDec(12_000_000)},
			expectPass:     true,
		},
		{
			name:      "redelegate to new set of validators",
			delegator: sdk.AccAddress([]byte("addr3---------------")),
			newPreferences: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[4],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[5],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[6],
					Weight:         sdk.NewDecWithPrec(2, 1),
				},
				{
					ValOperAddress: valAddrs[7],
					Weight:         sdk.NewDecWithPrec(1, 1),
				},
				{
					ValOperAddress: valAddrs[8],
					Weight:         sdk.NewDecWithPrec(3, 1),
				},
			},
			coinToStake:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50_000_000)),
			expectedShares: []sdk.Dec{sdk.NewDec(10_000_000), sdk.NewDec(10_000_000), sdk.NewDec(10_000_000), sdk.NewDec(5_000_000), sdk.NewDec(15_000_000)},
			expectPass:     true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {

			// fund the account that is trying to delegate
			suite.FundAcc(test.delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)})

			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			if !test.delegationExist {
				// creates a validator preference list to delegate to
				preferences := suite.PrepareDelegateToValidatorSet()

				// SetValidatorSetPreference sets a new list of val-set
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				suite.Require().NoError(err)

				// DelegateToValidatorSet delegate to existing val-set
				_, err = msgServer.DelegateToValidatorSet(c, types.NewMsgDelegateToValidatorSet(test.delegator, test.coinToStake))
				suite.Require().NoError(err)
			}

			// RedelegateValidatorSet redelegates from an exisitng set to a new one
			_, err := msgServer.RedelegateValidatorSet(c, types.NewMsgRedelegateValidatorSet(test.delegator, test.newPreferences))
			if test.expectPass {
				suite.Require().NoError(err)

				// check if the validator have recieved the correct amount of tokens
				for i, val := range test.newPreferences {
					valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
					suite.Require().NoError(err)

					// guarantees that the delegator exists because we check it in DelegateToValidatorSet
					del, _ := suite.App.StakingKeeper.GetDelegation(suite.Ctx, test.delegator, valAddr)
					suite.Require().Equal(del.Shares, test.expectedShares[i])
				}

			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite *KeeperTestSuite) TestWithdrawDelegationRewards() {
	suite.SetupTest()

	// call the create validator set preference
	preferences := suite.PrepareDelegateToValidatorSet()
	// setup extra validator for non val-set staking position
	valAddrs := suite.SetupMultipleValidators(1)

	amountToFund := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)}

	tests := []struct {
		name                  string
		delegator             sdk.AccAddress
		coinsToDelegate       sdk.Coin
		setValSetDelegation   bool
		setExistingDelegation bool
		expectPass            bool
	}{
		{
			name:                "Withdraw all rewards from existing valset delegations",
			delegator:           sdk.AccAddress([]byte("addr1---------------")),
			coinsToDelegate:     sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)), // delegate 20osmo
			setValSetDelegation: true,
			expectPass:          true,
		},
		{
			name:                  "Withdraw all rewards from existing staking delegations ",
			delegator:             sdk.AccAddress([]byte("addr2---------------")),
			coinsToDelegate:       sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)),
			setExistingDelegation: true,
			expectPass:            true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.FundAcc(test.delegator, amountToFund) // 100 osmo

			// setup message server
			msgServer := valPref.NewMsgServerImpl(suite.App.ValidatorSetPreferenceKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)

			ctx := suite.Ctx
			// setup test for only valset delegation
			if test.setValSetDelegation {
				// delegators have to set val-set before delegating tokens
				_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(test.delegator, preferences))
				suite.Require().NoError(err)

				// call the delegate to validator set preference message
				_, err = msgServer.DelegateToValidatorSet(c, types.NewMsgDelegateToValidatorSet(test.delegator, test.coinsToDelegate))
				suite.Require().NoError(err)

				suite.SetupDelegationReward(ctx, test.delegator, preferences, "", test.setValSetDelegation, test.setExistingDelegation)
			}

			// setup test for only existing staking position
			if test.setExistingDelegation {
				err := suite.PrepareExistingDelegations(suite.Ctx, valAddrs, test.delegator, test.coinsToDelegate.Amount)
				suite.Require().NoError(err)

				suite.SetupDelegationReward(ctx, test.delegator, nil, valAddrs[0], test.setValSetDelegation, test.setExistingDelegation)
			}

			_, err := msgServer.WithdrawDelegationRewards(c, types.NewMsgWithdrawDelegationRewards(test.delegator))
			if test.expectPass {
				suite.Require().NoError(err)

				// the rewards for valset and exising delegations should be nil
				if test.setValSetDelegation {
					for _, val := range preferences {
						rewardAfterWithdrawValSet, _ := suite.GetDelegationRewards(ctx, val.ValOperAddress, test.delegator)
						suite.Require().Nil(rewardAfterWithdrawValSet)
					}
				}

				if test.setExistingDelegation {
					rewardAfterWithdrawExistingSet, _ := suite.GetDelegationRewards(ctx, valAddrs[0], test.delegator)
					suite.Require().Nil(rewardAfterWithdrawExistingSet)
				}

			} else {
				suite.Require().Error(err)

			}
		})
	}
}
