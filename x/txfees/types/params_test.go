package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	appParams "github.com/osmosis-labs/osmosis/v31/app/params"
	"github.com/osmosis-labs/osmosis/v31/x/txfees/types"
)

const (
	validAddress      = "osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks"
	validIbcDenom     = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
	validFactoryDenom = "factory/osmo1z6r6qdknhgsc0zeracktgpcxf43j6sekq07nw8sxduc9lg0qjjlqfu25e3/alloyed/allBTC"
)

func TestParamsValidate(t *testing.T) {
	appParams.SetAddressPrefixes()

	testCases := map[string]struct {
		params    types.Params
		expectErr bool
	}{
		"default params": {
			params:    types.DefaultParams(),
			expectErr: false,
		},
		"valid custom lists": {
			params: types.Params{
				WhitelistedFeeTokenSetters:   []string{validAddress},
				FeeSwapIntermediaryDenomList: []string{"uosmo", validIbcDenom, validFactoryDenom},
			},
			expectErr: false,
		},
		"invalid address": {
			params: types.Params{
				WhitelistedFeeTokenSetters: []string{"cosmos1234"},
			},
			expectErr: true,
		},
		"invalid intermediary denom": {
			params: types.Params{
				FeeSwapIntermediaryDenomList: []string{"bad denom"},
			},
			expectErr: true,
		},
		"empty denom string": {
			params: types.Params{
				FeeSwapIntermediaryDenomList: []string{""},
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.params.Validate()

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
