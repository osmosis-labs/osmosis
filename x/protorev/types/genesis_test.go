package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

func TestGenesisStateValidate(t *testing.T) {
	trade1 := types.NewTrade(1, "a", "b")
	trade2 := types.NewTrade(2, "b", "c")
	routes := types.NewRoutes([]*types.Trade{&trade1, &trade2})

	invalidSearchRoutes := []types.TokenPairArbRoutes{
		types.NewTokenPairArbRoutes([]*types.Route{&routes}, "a", "b"),
	}

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
				Params:     types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination)},
			},
			valid: true,
		},
		{
			description: "Default parameters with invalid routes (duplicate token pairs)",
			genState: &types.GenesisState{
				Params:     types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination), types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination)},
			},
			valid: false,
		},
		{
			description: "Default parameters with nil routes",
			genState: &types.GenesisState{
				Params:     types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{types.NewTokenPairArbRoutes(nil, "a", "b")},
			},
			valid: false,
		},
		{
			description: "Default parameters with invalid routes (too few trades in a route)",
			genState: &types.GenesisState{
				Params:     types.DefaultParams(),
				TokenPairs: invalidSearchRoutes,
			},
			valid: false,
		},
		{
			description: "Default parameters with invalid routes (mismatch in and out denoms)",
			genState: &types.GenesisState{
				Params:     types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.OsmosisDenomination)},
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
