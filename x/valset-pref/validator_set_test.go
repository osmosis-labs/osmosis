package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/valset-pref/types"
)

func (suite *KeeperTestSuite) TestValidateLockForForceUnlock() {

	locks := suite.SetupLocks(sdk.AccAddress([]byte("addr1---------------")))

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
		suite.Run(test.name, func() {
			_, _, err := suite.App.ValidatorSetPreferenceKeeper.ValidateLockForForceUnlock(suite.Ctx, test.lockID, test.delegatorAddr)
			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCheckUndelegateTotalAmount() {
	valAddrs := suite.SetupMultipleValidators(3)
	tests := []struct {
		name        string
		tokenAmt    sdk.Dec
		existingSet []types.ValidatorPreference
		expectPass  bool
	}{
		{
			name:     "token amount matches with totalAmountFromWeights",
			tokenAmt: sdk.NewDec(122_312_231),
			existingSet: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(17, 2), // 0.17
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(83, 2), // 0.83
				},
			},
			expectPass: true,
		},
		{
			name:     "token decimal amount matches with totalAmountFromWeights",
			tokenAmt: sdk.MustNewDecFromStr("122312231.532"),
			existingSet: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(17, 2), // 0.17
				},
				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(83, 2), // 0.83
				},
			},
			expectPass: true,
		},
		{
			name:     "tokenAmt doesnot match with totalAmountFromWeights",
			tokenAmt: sdk.NewDec(122_312_231),
			existingSet: []types.ValidatorPreference{
				{
					ValOperAddress: valAddrs[0],
					Weight:         sdk.NewDecWithPrec(17, 2), // 0.17
				},

				{
					ValOperAddress: valAddrs[1],
					Weight:         sdk.NewDecWithPrec(83, 2), // 0.83
				},
				{
					ValOperAddress: valAddrs[2],
					Weight:         sdk.NewDecWithPrec(83, 2), // 0.83
				},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			err := suite.App.ValidatorSetPreferenceKeeper.CheckUndelegateTotalAmount(test.tokenAmt, test.existingSet)
			if test.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestIsValidatorSetEqual() {
	valAddrs := suite.SetupMultipleValidators(3)
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
		suite.Run(test.name, func() {
			isEqual := suite.App.ValidatorSetPreferenceKeeper.IsValidatorSetEqual(test.newValPreference, test.existingPreference)
			suite.Require().Equal(test.expectEqual, isEqual)
		})
	}

}

func (suite *KeeperTestSuite) TestIsPreferenceValid() {
	valAddrs := suite.SetupMultipleValidators(4)

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
		suite.Run(test.name, func() {
			valSet, err := suite.App.ValidatorSetPreferenceKeeper.IsPreferenceValid(suite.Ctx, test.valSetPreference)
			if test.expectPass {
				suite.Require().NoError(err)
				for i, vals := range valSet {
					suite.Require().Equal(test.expectedWeights[i], vals.Weight)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
