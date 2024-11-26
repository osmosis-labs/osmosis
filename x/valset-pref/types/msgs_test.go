package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

func TestMsgSetValidatorSetPreference(t *testing.T) {
	appParams.SetAddressPrefixes()
	addr1, invalidAddr := apptesting.GenerateTestAddrs()

	tests := []struct {
		name       string
		msg        types.MsgSetValidatorSetPreference
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgSetValidatorSetPreference{
				Delegator: addr1,
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDecWithPrec(322, 3),
					},
					{
						ValOperAddress: "osmovaloper1jcr68jghzm24zwe78zuhz7xahua8429erxk7vm",
						Weight:         osmomath.NewDecWithPrec(332, 3),
					},
					{
						ValOperAddress: "osmovaloper1gqsr38e4zteekwr6kq5se5jpadafqmcfyz8jds",
						Weight:         osmomath.NewDecWithPrec(348, 3),
					},
				},
			},
			expectPass: true,
		},
		{
			name: "duplicate validator msg",
			msg: types.MsgSetValidatorSetPreference{
				Delegator: addr1,
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDecWithPrec(6, 1),
					},
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDecWithPrec(4, 1),
					},
					{
						ValOperAddress: "osmovaloper1jcr68jghzm24zwe78zuhz7xahua8429erxk7vm",
						Weight:         osmomath.NewDecWithPrec(2, 1),
					},
				},
			},
			expectPass: false,
		},
		{
			name: "invalid delegator",
			msg: types.MsgSetValidatorSetPreference{
				Delegator: invalidAddr,
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDec(1),
					},
				},
			},
			expectPass: false,
		},
		{
			name: "invalid validator address",
			msg: types.MsgSetValidatorSetPreference{
				Delegator: addr1,
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxay", // invalid address
						Weight:         osmomath.NewDecWithPrec(2, 1),
					},
					{
						ValOperAddress: "osmovaloper1jcr68jghzm24zwe78zuhz7xahua8429erxk7vm",
						Weight:         osmomath.NewDecWithPrec(2, 1),
					},
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDecWithPrec(6, 1),
					},
				},
			},
			expectPass: false,
		},
		{
			name: "weights > 1",
			msg: types.MsgSetValidatorSetPreference{
				Delegator: addr1,
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDecWithPrec(5, 1),
					},
					{
						ValOperAddress: "osmovaloper1jcr68jghzm24zwe78zuhz7xahua8429erxk7vm",
						Weight:         osmomath.NewDecWithPrec(3, 1),
					},
					{
						ValOperAddress: "osmovaloper1gqsr38e4zteekwr6kq5se5jpadafqmcfyz8jds",
						Weight:         osmomath.NewDecWithPrec(3, 1),
					},
				},
			},
			expectPass: false,
		},
		{
			name: "weights < 1",
			msg: types.MsgSetValidatorSetPreference{
				Delegator: addr1,
				Preferences: []types.ValidatorPreference{
					{
						ValOperAddress: "osmovaloper1x2cfenmflhj3dwm2ph6nkgqr3nppkg86fxaymg",
						Weight:         osmomath.NewDecWithPrec(2, 1),
					},
					{
						ValOperAddress: "osmovaloper1jcr68jghzm24zwe78zuhz7xahua8429erxk7vm",
						Weight:         osmomath.NewDecWithPrec(2, 1),
					},
					{
						ValOperAddress: "osmovaloper1gqsr38e4zteekwr6kq5se5jpadafqmcfyz8jds",
						Weight:         osmomath.NewDecWithPrec(2, 1),
					},
				},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Type(), "set_validator_set_preference")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), addr1)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}
