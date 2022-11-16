package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

func TestGenesisState_Validate(t *testing.T) {
	cases := []struct {
		description string
		genState    *types.GenesisState
		valid       bool
	}{
		{
			description: "Default parameters with no routes",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
			},
			valid: true,
		},
		{
			description: "Default parameters with valid routes",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Routes: []types.Route{
					createRoute(types.AtomDenomination, 10),
					createRoute(types.AtomDenomination, 5),
					createRoute(types.OsmosisDenomination, 4),
				},
			},
			valid: true,
		},
		{
			description: "Default parameters with invalid routes",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Routes: []types.Route{
					createRoute(types.AtomDenomination, 10),
					createRoute(types.AtomDenomination, 1),
				},
			},
			valid: false,
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

func createRoute(arbDenom string, numberPools uint64) types.Route {
	pools := make([]uint64, numberPools)

	var pool uint64
	for pool = 0; pool < numberPools; pool++ {
		pools[pool] = pool
	}

	return types.Route{
		ArbDenom: arbDenom,
		Pools:    pools,
	}
}
