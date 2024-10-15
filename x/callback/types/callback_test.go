package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func TestCallbackValidate(t *testing.T) {
	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	validCoin := sdk.NewInt64Coin("stake", 1)

	type testCase struct {
		name        string
		callback    types.Callback
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "Fail: Empty values",
			callback:    types.Callback{},
			errExpected: true,
		},
		{
			name: "Fail: Invalid contract address",
			callback: types.Callback{
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
			errExpected: true,
		},
		{
			name: "Fail: Invalid reservedby address",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				ReservedBy:      "👻",
				CallbackHeight:  1,
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: Invalid callback height",
			callback: types.Callback{
				ContractAddress: contractAddr.String(),
				ReservedBy:      accAddr.String(),
				CallbackHeight:  -1,
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
			errExpected: true,
		},
		{
			name: "OK: Valid callback",
			callback: types.Callback{
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
			errExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.callback.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
