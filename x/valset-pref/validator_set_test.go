package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/valset-pref/types"
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
	tests := []struct {
		name        string
		tokenAmt    sdk.Dec
		existingSet []types.ValidatorPreference
		expectPass  bool
	}{
		{
			name:     "token Amount matches with totalAmountFromWeights",
			tokenAmt: sdk.NewDec(122_312_231),
			existingSet: []types.ValidatorPreference{
				{
					ValOperAddress: "addr1---------------",
					Weight:         sdk.NewDecWithPrec(17, 2), // 0.17
				},
				{
					ValOperAddress: "addr2---------------",
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
					ValOperAddress: "addr1---------------",
					Weight:         sdk.NewDecWithPrec(17, 2), // 0.17
				},

				{
					ValOperAddress: "addr2---------------",
					Weight:         sdk.NewDecWithPrec(83, 2), // 0.83
				},
				{
					ValOperAddress: "addr3---------------",
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
