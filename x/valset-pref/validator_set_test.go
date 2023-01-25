package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
