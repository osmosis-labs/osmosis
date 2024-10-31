package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

func TestGenesisStateValidate(t *testing.T) {
	cases := []struct {
		description string
		genState    *types.GenesisState
		valid       bool
	}{
		{
			description: "Default parameters with no routes",
			genState:    types.DefaultGenesis(),
			valid:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.genState.Validate()

			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestNullAccount(t *testing.T) {
	strAddress, err := sdk.AccAddressFromBech32("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030")
	require.NoError(t, err)
	require.True(t, types.DefaultNullAddress.Equals(strAddress))
}
