package types_test

import (
	sdkmath "cosmossdk.io/math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func TestParamsValidate(t *testing.T) {
	type testCase struct {
		name        string
		params      types.Params
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "OK: Default values",
			params:      types.DefaultParams(),
			errExpected: false,
		},
		{
			name: "OK: All valid values",
			params: types.NewParams(
				100,
				100,
				100,
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.ZeroInt()),
			),
			errExpected: false,
		},
		{
			name: "Fail: CallbackGasLimit: zero",
			params: types.NewParams(
				0,
				100,
				100,
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.ZeroInt()),
			),
			errExpected: true,
		},
		{
			name: "Fail: BlockReservationFeeMultiplier: negative",
			params: types.NewParams(
				100,
				100,
				100,
				sdkmath.LegacyMustNewDecFromStr("-1.0"),
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.ZeroInt()),
			),
			errExpected: true,
		},
		{
			name: "Fail: FutureReservationFeeMultiplier: negative",
			params: types.NewParams(
				100,
				100,
				100,
				sdkmath.LegacyMustNewDecFromStr("1.0"),
				sdkmath.LegacyMustNewDecFromStr("-1.0"),
				sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.ZeroInt()),
			),
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
