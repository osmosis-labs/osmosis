package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func TestGenesisValidate(t *testing.T) {
	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	validCoin := sdk.NewInt64Coin("stake", 1)

	type testCase struct {
		name        string
		genesis     types.GenesisState
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "Fail: Empty values",
			genesis:     types.GenesisState{},
			errExpected: true,
		},
		{
			name: "Fail: Invalid params",
			genesis: types.GenesisState{
				Params: types.NewParams(
					0,
					100,
					100,
					math.LegacyMustNewDecFromStr("1.0"),
					math.LegacyMustNewDecFromStr("1.0"),
					sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
				),
				Callbacks: []*types.Callback{
					{
						ContractAddress: contractAddr.String(),
						ReservedBy:      accAddr.String(),
						CallbackHeight:  1,
						FeeSplit: &types.CallbackFeesFeeSplit{
							TransactionFees:       &validCoin,
							BlockReservationFees:  &validCoin,
							FutureReservationFees: &validCoin,
							SurplusFees:           &validCoin,
						},
					},
				},
			},
			errExpected: true,
		}, {
			name: "Fail: Invalid callback",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Callbacks: []*types.Callback{
					{
						ContractAddress: "👻",
						ReservedBy:      accAddr.String(),
						CallbackHeight:  1,
						FeeSplit: &types.CallbackFeesFeeSplit{
							TransactionFees:       &validCoin,
							BlockReservationFees:  &validCoin,
							FutureReservationFees: &validCoin,
							SurplusFees:           &validCoin,
						},
					},
				},
			},
			errExpected: true,
		},
		{
			name: "OK: Valid genesis state",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Callbacks: []*types.Callback{
					{
						ContractAddress: contractAddr.String(),
						ReservedBy:      accAddr.String(),
						CallbackHeight:  1,
						FeeSplit: &types.CallbackFeesFeeSplit{
							TransactionFees:       &validCoin,
							BlockReservationFees:  &validCoin,
							FutureReservationFees: &validCoin,
							SurplusFees:           &validCoin,
						},
					},
				},
			},
			errExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
