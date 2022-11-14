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
