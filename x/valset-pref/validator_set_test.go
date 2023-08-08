package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	valPref "github.com/osmosis-labs/osmosis/v17/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v17/x/valset-pref/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
			name:          "lock Id doesnot match with delegator",
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
			Weight:         sdk.NewDecWithPrec(5, 1),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.NewDecWithPrec(5, 1),
		},
	}

	valSetTwo := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         sdk.NewDecWithPrec(5, 1),
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         sdk.NewDecWithPrec(5, 1),
		},
	}

	valSetThree := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         sdk.NewDecWithPrec(2, 1),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.NewDecWithPrec(8, 1),
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

func (s *KeeperTestSuite) TestIsPreferenceValid() {
	valAddrs := s.SetupMultipleValidators(4)

	tests := []struct {
		name             string
		valSetPreference []types.ValidatorPreference
		expectedWeights  []sdk.Dec
		expectPass       bool
	}{
		{
			name: "Valid Prefernce: Check rounding",
			valSetPreference: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.MustNewDecFromStr("0.3315"), // rounds to = 0.33
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.MustNewDecFromStr("0.000"), // rounds to = 0
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.MustNewDecFromStr("0.536"), // rounds to = 0.54
				},
				{
					ValOperAddress: valAddrs[3],
					Weight:         sdk.MustNewDecFromStr("0.119"), // rounds to = 0.12
				},
			},
			expectedWeights: []sdk.Dec{sdk.NewDecWithPrec(33, 2), sdk.ZeroDec(), sdk.NewDecWithPrec(54, 2), sdk.NewDecWithPrec(12, 2)},
			expectPass:      true,
		},
		{
			name: "Invalid preference, invalid validator",
			valSetPreference: []types.ValidatorPreference{
				{
					ValOperAddress: "addr1---------------",
					Weight:         sdk.MustNewDecFromStr("0.3415"),
				},
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.MustNewDecFromStr("0.000"),
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

	valAddrs := s.SetupMultipleValidators(2)
	valPreferences := []types.ValidatorPreference{
		{
			ValOperAddress: valAddrs[0],
			Weight:         sdk.NewDecWithPrec(5, 1), // 0.5
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.NewDecWithPrec(5, 1), // 0.5
		},
	}

	delegator := sdk.AccAddress([]byte("addr1---------------"))
	coinToStake := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10_000_000))   // delegate 10osmo using Valset now and 10 osmo using regular staking delegate
	coinToUnStake := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20_000_000)) // undelegate 20osmo
	expectedShares := []sdk.Dec{sdk.NewDec(15_000_000), sdk.NewDec(500_000)}

	s.FundAcc(delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100_000_000)}) // 100 osmo

	// valset test setup
	// SetValidatorSetPreference sets a new list of val-set
	_, err := s.App.ValidatorSetPreferenceKeeper.SetValidatorSetPreference(s.Ctx, delegator.String(), valPreferences)
	s.Require().NoError(err)

	s.App.ValidatorSetPreferenceKeeper.SetValidatorSetPreferences(s.Ctx, delegator.String(), types.ValidatorSetPreferences{
		Preferences: valPreferences,
	})

	// DelegateToValidatorSet delegate to existing val-set
	err = s.App.ValidatorSetPreferenceKeeper.DelegateToValidatorSet(s.Ctx, delegator.String(), coinToStake)
	s.Require().NoError(err)

	valAddr, err := sdk.ValAddressFromBech32(valAddrs[0])
	s.Require().NoError(err)

	validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().True(found)

	// Delegate more token to the validator. This will cause valset and regular staking to go out of sync
	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, sdk.NewInt(10_000_000), stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)

	err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromValidatorSet(s.Ctx, delegator.String(), coinToUnStake)
	s.Require().NoError(err)

	for i, val := range valPreferences {
		valAddr, err := sdk.ValAddressFromBech32(val.ValOperAddress)
		s.Require().NoError(err)

		// guarantees that the delegator exists because we check it in UnDelegateToValidatorSet
		del, found := s.App.StakingKeeper.GetDelegation(s.Ctx, delegator, valAddr)
		if found {
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
			Weight:         sdk.MustNewDecFromStr("0.05"),
		},
		{
			ValOperAddress: valAddrs[1],
			Weight:         sdk.MustNewDecFromStr("0.05"),
		},
		{
			ValOperAddress: valAddrs[2],
			Weight:         sdk.NewDecWithPrec(45, 2), // 0.45
		},
		{
			ValOperAddress: valAddrs[3],
			Weight:         sdk.NewDecWithPrec(45, 2), // 0.45
		},
	}

	delegator := sdk.AccAddress([]byte("addr4---------------"))
	coinToStake := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100_000_000))   // delegate 100osmo using Valset now and 10 osmo using regular staking delegate
	coinToUnStake := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(200_000_000)) // undelegate 20osmo

	s.FundAcc(delegator, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 300_000_000)}) // 100 osmo

	// valset test setup
	// SetValidatorSetPreference sets a new list of val-set
	msgServer := valPref.NewMsgServerImpl(s.App.ValidatorSetPreferenceKeeper)
	c := sdk.WrapSDKContext(s.Ctx)

	// SetValidatorSetPreference sets a new list of val-set
	_, err := msgServer.SetValidatorSetPreference(c, types.NewMsgSetValidatorSetPreference(delegator, valPreferences))
	s.Require().NoError(err)

	// DelegateToValidatorSet delegate to existing val-set
	err = s.App.ValidatorSetPreferenceKeeper.DelegateToValidatorSet(s.Ctx, delegator.String(), coinToStake)
	s.Require().NoError(err)

	valAddr, err := sdk.ValAddressFromBech32(valAddrs[0])
	s.Require().NoError(err)

	validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().True(found)

	valAddr2, err := sdk.ValAddressFromBech32(valAddrs[1])
	s.Require().NoError(err)

	validator2, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr2)
	s.Require().True(found)

	// Delegate more token to the validator. This will cause valset and regular staking to go out of sync
	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, sdk.NewInt(50_000_000), stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err)

	_, err = s.App.StakingKeeper.Delegate(s.Ctx, delegator, sdk.NewInt(50_000_000), stakingtypes.Unbonded, validator2, true)
	s.Require().NoError(err)

	err = s.App.ValidatorSetPreferenceKeeper.UndelegateFromValidatorSet(s.Ctx, delegator.String(), coinToUnStake)
	s.Require().NoError(err)

}
