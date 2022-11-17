package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

func TestGenesisStateValidate(t *testing.T) {
	routes := []*types.Route{
		{
			Pools: []uint64{1, 2},
		},
	}
	invalidSearchRoutes := []types.SearcherRoutes{
		types.NewSearcherRoutes(routes, "a", "b"),
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
				Params: types.DefaultParams(),
				Routes: []types.SearcherRoutes{createSeacherRoutes(3, types.AtomDenomination, types.OsmosisDenomination)},
			},
			valid: true,
		},
		{
			description: "Default parameters with invalid routes (duplicate pools)",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Routes: []types.SearcherRoutes{createSeacherRoutes(3, types.AtomDenomination, types.OsmosisDenomination), createSeacherRoutes(3, types.AtomDenomination, types.OsmosisDenomination)},
			},
			valid: false,
		},
		{
			description: "Default parameters with nil routes",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Routes: []types.SearcherRoutes{types.NewSearcherRoutes(nil, "a", "b")},
			},
			valid: false,
		},
		{
			description: "Default parameters with invalid routes",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Routes: invalidSearchRoutes,
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

// CreateRoute creates SearchRoutes object for testing
func createSeacherRoutes(numberPools uint64, tokenA, tokenB string) types.SearcherRoutes {
	routes := make([]*types.Route, numberPools)
	for i := uint64(0); i < numberPools; i++ {
		routes[i] = &types.Route{
			Pools: []uint64{i, i + 1, i + 2},
		}
	}

	return types.NewSearcherRoutes(routes, tokenA, tokenB)
}
