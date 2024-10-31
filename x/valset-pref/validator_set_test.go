package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	valPref "github.com/osmosis-labs/osmosis/v27/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

func (s *KeeperTestSuite) TestValidateLockForForceUnlock() {
	locks := s.SetupLocks(sdk.AccAddress([]byte("addr1---------------")))

	tests := []struct {
		name          string
		lockID        uint64
		delegatorAddr string
		expectPass    bool
	}{
		{
			name:          "happy case",
			lockID:        locks[0].ID,
			delegatorAddr: sdk.AccAddress([]byte("addr1---------------")).String(),
			expectPass:    true,
		},
		{
			name:          "lock Id does not match with delegator",
			lockID:        locks[0].ID,
			delegatorAddr: "addr2---------------",
			expectPass:    false,
		},
		{
			name:          "Invalid Lock: contains multiple coins",
			lockID:        locks[6].ID,
			delegatorAddr: "addr1---------------",
			expectPass:    false,
		},
		{
			name:          "Invalid Lock: contains non osmo denom",
			lockID:        locks[1].ID,
			delegatorAddr: "addr1---------------",
			expectPass:    false,
		},
		{
			name:          "Invalid Lock: contains lock with duration > 2 weeks",
			lockID:        locks[3].ID,
			delegatorAddr: "addr1---------------",
			expectPass:    false,
		},
		{
			name:          "Invalid lock: non bonded lockId",
			lockID:        locks[4].ID,
			delegatorAddr: "addr1---------------",
			expectPass:    false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			_, _, err := s.App.ValidatorSetPreferenceKeeper.ValidateLockForForceUnlock(s.Ctx, test.lockID, test.delegatorAddr)
			if test.expectPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestIsValidatorSetEqual() {
	valAddrs := s.SetupMultipleValidators(3)
	valSetOne := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.NewDecWithPrec(5, 1),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         osmomath.NewDecWithPrec(5, 1),
		},
	}

	valSetTwo := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.NewDecWithPrec(5, 1),
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         osmomath.NewDecWithPrec(5, 1),
		},
	}

	valSetThree := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.NewDecWithPrec(2, 1),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         osmomath.NewDecWithPrec(8, 1),
		},
	}

	tests := []struct {
		name               string
		existingPreference []types.ValidatorPreference
		newValPreference   []types.ValidatorPreference
		expectEqual        bool
	}{
		{
			name:               "Valsets: same address and same weights",
			existingPreference: valSetOne,
			newValPreference:   valSetOne,
			expectEqual:        true,
		},
		{
			name:               "Valsets: same address, different weights",
			existingPreference: valSetOne,
			newValPreference:   valSetThree,
			expectEqual:        false,
		},
		{
			name:               "ValSets: different address, same weights",
			existingPreference: valSetOne,
			newValPreference:   valSetTwo,
			expectEqual:        false,
		},
		{
			name:               "ValSets: different address, different weights",
			existingPreference: valSetOne,
			newValPreference:   valSetThree,
			expectEqual:        false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			isEqual := s.App.ValidatorSetPreferenceKeeper.IsValidatorSetEqual(test.newValPreference, test.existingPreference)
			s.Require().Equal(test.expectEqual, isEqual)
		})
	}
}

func (s *KeeperTestSuite) TestUndelegateFromValidatorSet() {
	tests := []struct {
		name                  string
		delegateAmt           []osmomath.Int
		undelegateAmt         osmomath.Int
		noValset              bool
		expectedUndelegateAmt []osmomath.Int
		expectedError         error
	}{
		{
			name:                  "exit at step 4: undelegating amount is under existing delegation amount",
			delegateAmt:           []osmomath.Int{osmomath.NewInt(100), osmomath.NewInt(50)},
			undelegateAmt:         osmomath.NewInt(50),
			expectedUndelegateAmt: []osmomath.Int{osmomath.NewInt(33), osmomath.NewInt(17)},
		},
		{
			name:          "error: attempt to undelegate more than delegated",
			delegateAmt:   []osmomath.Int{osmomath.NewInt(100), osmomath.NewInt(50)},
			undelegateAmt: osmomath.NewInt(200),
			expectedError: types.UndelegateMoreThanDelegatedError{TotalDelegatedAmt: osmomath.NewDec(150), UndelegationAmt: osmomath.NewInt(200)},
		},
		{
			name:          "error: user does not have val-set preference set",
			delegateAmt:   []osmomath.Int{osmomath.NewInt(100), osmomath.NewInt(50)},
			undelegateAmt: osmomath.NewInt(100),
			noValset:      true,
			expectedError: types.NoValidatorSetOrExistingDelegationsError{DelegatorAddr: s.TestAccs[0].String()},
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			valAddrs := s.SetupMultipleValidators(3)
			defaultDelegator := s.TestAccs[0]
			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)

			// set val-set pref
			valPreferences := []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         osmomath.NewDecWithPrec(1, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         osmomath.NewDecWithPrec(9, 1),
				},
			}

			if !test.noValset {
				s.App.ValidatorSetPreferenceKeeper.SetValidatorSetPreferences(s.Ctx, defaultDelegator.String(), types.ValidatorSetPreferences{
					Preferences: valPreferences,
				})
				// delegate for each of the validators
				for i, valsetPref := range valPreferences {
					valAddr, err := sdk.ValAddressFromBech32(valsetPref.ValOperAddress)
					s.Require().NoError(err)
					validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
					s.Require().NoError(err)

					s.FundAcc(defaultDelegator, sdk.NewCoins(sdk.NewCoin(bondDenom, test.delegateAmt[i])))
					_, err = s.App.StakingKeeper.Delegate(s.Ctx, defaultDelegator, test.delegateAmt[i], stakingtypes.Unbonded, validator, true)
					s.Require().NoError(err)
				}
			}

			// System Under Test
			err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromValidatorSet(s.Ctx, defaultDelegator.String(), sdk.NewCoin(bondDenom, test.undelegateAmt))

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedError.Error())
				return
			}
			s.Require().NoError(err)

			for i, valsetPref := range valPreferences {
				valAddr, err := sdk.ValAddressFromBech32(valsetPref.ValOperAddress)
				s.Require().NoError(err)

				delegation, err := s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, defaultDelegator, valAddr)
				s.Require().NoError(err)
				s.Require().Equal(delegation.Entries[0].Balance, test.expectedUndelegateAmt[i])
			}
		})
	}
}

func (s *KeeperTestSuite) TestUndelegateFromRebalancedValidatorSet() {
	tests := []struct {
		name                  string
		delegateAmt           []osmomath.Int
		undelegateAmt         osmomath.Int
		noValset              bool
		expectedUndelegateAmt []osmomath.Int
		expectedError         error
	}{
		{
			name:                  "happy path: undelegate all, weights match the current delegations to valset",
			delegateAmt:           []osmomath.Int{osmomath.NewInt(10), osmomath.NewInt(90)},
			undelegateAmt:         osmomath.NewInt(100),
			expectedUndelegateAmt: []osmomath.Int{osmomath.NewInt(10), osmomath.NewInt(90)},
		},
		{
			name:                  "happy path: undelegate some, weights match the current delegations to valset",
			delegateAmt:           []osmomath.Int{osmomath.NewInt(10), osmomath.NewInt(90)},
			undelegateAmt:         osmomath.NewInt(50),
			expectedUndelegateAmt: []osmomath.Int{osmomath.NewInt(5), osmomath.NewInt(45)},
		},
		{
			name:                  "undelegate all, weights do not match the current delegations to valset",
			delegateAmt:           []osmomath.Int{osmomath.NewInt(90), osmomath.NewInt(10)},
			undelegateAmt:         osmomath.NewInt(100),
			expectedUndelegateAmt: []osmomath.Int{osmomath.NewInt(90), osmomath.NewInt(10)},
		},
		{
			name:                  "undelegate some, weights do not match the current delegations to valset",
			delegateAmt:           []osmomath.Int{osmomath.NewInt(90), osmomath.NewInt(10)},
			undelegateAmt:         osmomath.NewInt(50),
			expectedUndelegateAmt: []osmomath.Int{osmomath.NewInt(45), osmomath.NewInt(5)},
		},
		{
			name:          "error: attempt to undelegate more than delegated",
			delegateAmt:   []osmomath.Int{osmomath.NewInt(100), osmomath.NewInt(50)},
			undelegateAmt: osmomath.NewInt(200),
			expectedError: types.UndelegateMoreThanDelegatedError{TotalDelegatedAmt: osmomath.NewDec(150), UndelegationAmt: osmomath.NewInt(200)},
		},
		{
			name:          "error: user does not have val-set preference set",
			delegateAmt:   []osmomath.Int{osmomath.NewInt(100), osmomath.NewInt(50)},
			undelegateAmt: osmomath.NewInt(100),
			noValset:      true,
			expectedError: types.NoValidatorSetOrExistingDelegationsError{DelegatorAddr: s.TestAccs[0].String()},
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			valAddrs := s.SetupMultipleValidators(3)
			defaultDelegator := s.TestAccs[0]
			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)

			// set val-set pref
			valPreferences := []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         osmomath.NewDecWithPrec(1, 1),
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         osmomath.NewDecWithPrec(9, 1),
				},
			}

			if !test.noValset {
				s.App.ValidatorSetPreferenceKeeper.SetValidatorSetPreferences(s.Ctx, defaultDelegator.String(), types.ValidatorSetPreferences{
					Preferences: valPreferences,
				})
				// delegate for each of the validators
				for i, valsetPref := range valPreferences {
					valAddr, err := sdk.ValAddressFromBech32(valsetPref.ValOperAddress)
					s.Require().NoError(err)
					validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
					s.Require().NoError(err)

					s.FundAcc(defaultDelegator, sdk.NewCoins(sdk.NewCoin(bondDenom, test.delegateAmt[i])))
					_, err = s.App.StakingKeeper.Delegate(s.Ctx, defaultDelegator, test.delegateAmt[i], stakingtypes.Unbonded, validator, true)
					s.Require().NoError(err)
				}
			}

			// System Under Test
			err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromRebalancedValidatorSet(s.Ctx, defaultDelegator.String(), sdk.NewCoin(bondDenom, test.undelegateAmt))

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorContains(err, test.expectedError.Error())
				return
			}
			s.Require().NoError(err)

			for i, valsetPref := range valPreferences {
				valAddr, err := sdk.ValAddressFromBech32(valsetPref.ValOperAddress)
				s.Require().NoError(err)

				delegation, err := s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, defaultDelegator, valAddr)
				s.Require().NoError(err)
				s.Require().Equal(delegation.Entries[0].Balance, test.expectedUndelegateAmt[i])
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetValsetRatios() {
	defaultDelegationAmt := osmomath.NewInt(100)
	tests := []struct {
		name              string
		useSingleValPref  bool
		undelegateAmt     osmomath.Int
		expectedError     bool
		notDelegated      bool
		expectedValRatios []valPref.ValRatio
	}{
		{
			name:             "single validator, undelegate full amount",
			useSingleValPref: true,
			undelegateAmt:    osmomath.NewInt(100),
			expectedValRatios: []valPref.ValRatio{
				{
					Weight:        osmomath.NewDec(1),
					DelegatedAmt:  defaultDelegationAmt,
					UndelegateAmt: defaultDelegationAmt,
					VRatio:        osmomath.NewDec(1),
				},
			},
		},
		{
			name:             "single validator, undelegate partial amount",
			useSingleValPref: true,
			undelegateAmt:    osmomath.NewInt(50),
			expectedValRatios: []valPref.ValRatio{
				{
					Weight:        osmomath.NewDec(1),
					DelegatedAmt:  defaultDelegationAmt,
					UndelegateAmt: defaultDelegationAmt.Quo(osmomath.NewInt(2)),
					// 0.5 since we are undelegating half amount
					VRatio: osmomath.NewDecWithPrec(5, 1),
				},
			},
		},
		{
			name:          "multiple validator, undelegate full amount",
			undelegateAmt: defaultDelegationAmt,
			expectedValRatios: []valPref.ValRatio{
				{
					Weight:        osmomath.MustNewDecFromStr("0.333333333333333333"),
					DelegatedAmt:  defaultDelegationAmt,
					UndelegateAmt: osmomath.NewInt(33),
					VRatio:        osmomath.MustNewDecFromStr("0.33"),
				},
				{
					Weight:        osmomath.MustNewDecFromStr("0.666666666666666667"),
					DelegatedAmt:  defaultDelegationAmt,
					UndelegateAmt: osmomath.NewInt(66),
					VRatio:        osmomath.MustNewDecFromStr("0.66"),
				},
			},
		},
		{
			name:          "multiple validator, undelegate partial amount",
			undelegateAmt: defaultDelegationAmt.Quo(osmomath.NewInt(2)),
			expectedValRatios: []valPref.ValRatio{
				{
					Weight:       osmomath.MustNewDecFromStr("0.333333333333333333"),
					DelegatedAmt: defaultDelegationAmt,
					// 1/3 of undelegating amount(50)
					UndelegateAmt: osmomath.NewInt(16),
					VRatio:        osmomath.MustNewDecFromStr("0.16"),
				},
				{
					Weight:       osmomath.MustNewDecFromStr("0.666666666666666667"),
					DelegatedAmt: defaultDelegationAmt,
					// 2/3 of undelegating amount(50)
					UndelegateAmt: osmomath.NewInt(33),
					VRatio:        osmomath.MustNewDecFromStr("0.33"),
				},
			},
		},
		{
			name:             "error: not delegated",
			undelegateAmt:    defaultDelegationAmt,
			useSingleValPref: true,
			notDelegated:     true,
			expectedError:    true,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			valAddrs := s.SetupMultipleValidators(3)
			defaultDelegator := s.TestAccs[0]

			var valsetPrefs []types.ValidatorPreference
			if test.useSingleValPref {
				valsetPrefs = []types.ValidatorPreference{
					{
						ValOperAddress: valAddrs[0],
						Weight:         osmomath.OneDec(),
					},
				}
			} else { // other cases, we assume we are using val set pref with multiple validators
				valsetPrefs = []types.ValidatorPreference{
					{
						ValOperAddress: valAddrs[0],
						Weight:         osmomath.MustNewDecFromStr("0.333333333333333333"),
					},
					{
						ValOperAddress: valAddrs[1],
						Weight:         osmomath.MustNewDecFromStr("0.666666666666666667"),
					},
				}
			}

			// set up delegation for each of the valset prefs
			expectedTotalDelegatedAmt := osmomath.ZeroDec()
			if !test.notDelegated {
				for i, valsetPref := range valsetPrefs {
					valAddr, err := sdk.ValAddressFromBech32(valsetPref.ValOperAddress)
					s.Require().NoError(err)
					validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
					s.Require().NoError(err)
					bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
					s.Require().NoError(err)

					s.FundAcc(defaultDelegator, sdk.NewCoins(sdk.NewCoin(bondDenom, defaultDelegationAmt)))
					_, err = s.App.StakingKeeper.Delegate(s.Ctx, defaultDelegator, defaultDelegationAmt, stakingtypes.Unbonded, validator, true)
					s.Require().NoError(err)

					expectedTotalDelegatedAmt = expectedTotalDelegatedAmt.Add(defaultDelegationAmt.ToLegacyDec())
					test.expectedValRatios[i].ValAddr = valAddr
				}
			}

			// system under test
			valRatios, validators, totalDelegatedAmt, err := s.App.ValidatorSetPreferenceKeeper.GetValsetRatios(s.Ctx, defaultDelegator, valsetPrefs, test.undelegateAmt)
			if test.expectedError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().True(totalDelegatedAmt.Equal(expectedTotalDelegatedAmt))
			// iterate over returned validators, make sure correct validators are returned in the map
			for valAddr, val := range validators {
				valAddr, err := sdk.ValAddressFromBech32(valAddr)
				s.Require().NoError(err)
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
				s.Require().NoError(err)
				validator.Equal(&val)
			}

			s.Require().Equal(valRatios, test.expectedValRatios)
		})
	}
}

func (s *KeeperTestSuite) TestIsPreferenceValid() {
	valAddrs := s.SetupMultipleValidators(4)

	tests := []struct {
		name             string
		valSetPreference []types.ValidatorPreference
		expectedWeights  []osmomath.Dec
		expectPass       bool
	}{
		{
			name: "Valid Preference: Check rounding",
			valSetPreference: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         osmomath.MustNewDecFromStr("0.3315"), // rounds to = 0.33
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         osmomath.MustNewDecFromStr("0.000"), // rounds to = 0
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         osmomath.MustNewDecFromStr("0.536"), // rounds to = 0.54
				},
				{
					ValOperAddress: valAddrs[3],
					Weight:         osmomath.MustNewDecFromStr("0.119"), // rounds to = 0.12
				},
			},
			expectedWeights: []osmomath.Dec{osmomath.NewDecWithPrec(33, 2), osmomath.ZeroDec(), osmomath.NewDecWithPrec(54, 2), osmomath.NewDecWithPrec(12, 2)},
			expectPass:      true,
		},
		{
			name: "Invalid preference, invalid validator",
			valSetPreference: []types.ValidatorPreference{
				{
					ValOperAddress: "addr1---------------",
					Weight:         osmomath.MustNewDecFromStr("0.3415"),
				},
				{
					ValOperAddress: valAddrs[0],
					Weight:         osmomath.MustNewDecFromStr("0.000"),
				},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			valSet, err := s.App.ValidatorSetPreferenceKeeper.IsPreferenceValid(s.Ctx, test.valSetPreference)
			if test.expectPass {
				s.Require().NoError(err)
				for i, vals := range valSet {
					s.Require().Equal(test.expectedWeights[i], vals.Weight)
				}
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// NOTE: this is the case that used to error. Fixed by this PR
func (s *KeeperTestSuite) TestUndelegateFromValSetErrorCase() {
	s.SetupTest()

	valAddrs := s.SetupMultipleValidators(3)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.NewDecWithPrec(5, 1), // 0.5
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         osmomath.NewDecWithPrec(5, 1), // 0.5
		},
	}

	delegator := sdk.AccAddress([]byte("addr1---------------"))
	coinToStake := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(10_000_000))   // delegate 10osmo using Valset now and 10 osmo using regular staking delegate
	coinToUnStake := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(20_000_000)) // undelegate 20osmo
	expectedShares := []osmomath.Dec{osmomath.NewDec(15_000_000), osmomath.NewDec(500_000)}

	s.FundAcc(delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)}) // 100 osmo

	// valset test setup
	// SetValidatorSetPreference sets a new list of val-set
	_, err := s.App.ValidatorSetPreferenceKeeper.ValidateValidatorSetPreference(s.Ctx, delegator.String(), valPreferences)
	s.Require().NoError(err)

	s.App.ValidatorSetPreferenceKeeper.SetValidatorSetPreferences(s.Ctx, delegator.String(), types.ValidatorSetPreferences{
		Preferences: valPreferences,
	})

	// DelegateToValidatorSet delegate to existing val-set
	err = s.App.ValidatorSetPreferenceKeeper.DelegateToValidatorSet(s.Ctx, delegator.String(), coinToStake)
	s.Require().NoError(err)

	valAddr, err := sdk.ValAddressFromBech32(valAddrs[0])
	s.Require().NoError(err)

	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	// Delegate more token to the validator. This will cause valset and regular staking to go out of sync
	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, osmomath.NewInt(10_000_000), stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)

	err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromValidatorSet(s.Ctx, delegator.String(), coinToUnStake)
	s.Require().NoError(err)

	for i, val := range valPreferences {
		valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
		s.Require().NoError(err)

		// guarantees that the delegator exists because we check it in UnDelegateToValidatorSet
		del, err := s.App.StakingKeeper.GetDelegation(s.Ctx, delegator, valAddr)
		if err == nil {
			s.Require().Equal(expectedShares[i], del.GetShares())
		}
	}

}

// NOTE: this is the case that used to error. Fixed by this PR
func (s *KeeperTestSuite) TestUndelegateFromValSetErrorCase1() {
	s.SetupTest()

	valAddrs := s.SetupMultipleValidators(4)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         osmomath.MustNewDecFromStr("0.05"),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         osmomath.MustNewDecFromStr("0.05"),
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         osmomath.NewDecWithPrec(45, 2), // 0.45
		},
		{
			ValOperAddress: valAddrs[3],
			Weight:         osmomath.NewDecWithPrec(45, 2), // 0.45
		},
	}

	delegator := sdk.AccAddress([]byte("addr4---------------"))
	coinToStake := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(100_000_000))   // delegate 100osmo using Valset now and 10 osmo using regular staking delegate
	coinToUnStake := sdk.NewCoin(sdk.DefaultBondDenom, osmomath.NewInt(200_000_000)) // undelegate 20osmo

	s.FundAcc(delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 300_000_000)}) // 100 osmo

	// valset test setup
	// SetValidatorSetPreference sets a new list of val-set
	msgServer := valPref.NewMsgServerImpl(s.App.ValidatorSetPreferenceKeeper)
	c := s.Ctx

	// SetValidatorSetPreference sets a new list of val-set
	_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(delegator, valPreferences))
	s.Require().NoError(err)

	// DelegateToValidatorSet delegate to existing val-set
	err = s.App.ValidatorSetPreferenceKeeper.DelegateToValidatorSet(s.Ctx, delegator.String(), coinToStake)
	s.Require().NoError(err)

	valAddr, err := sdk.ValAddressFromBech32(valAddrs[0])
	s.Require().NoError(err)

	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	valAddr2, err := sdk.ValAddressFromBech32(valAddrs[1])
	s.Require().NoError(err)

	validator2, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr2)
	s.Require().NoError(err)

	// Delegate more token to the validator. This will cause valset and regular staking to go out of sync
	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, osmomath.NewInt(50_000_000), stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)

	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, osmomath.NewInt(50_000_000), stakingtypes.Unbonded, validator2, true)
	s.Require().NoError(err)

	err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromValidatorSet(s.Ctx, delegator.String(), coinToUnStake)
	s.Require().NoError(err)

}
