package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/tokenfactory/types"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "different admin from creator",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "symphony167psfumv4erluc3ksfw23n2pdrx6zx3rj2guvl",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "empty admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "no admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
					},
				},
			},
			valid: true,
		},
		{
			desc: "invalid admin",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "moose",
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "multiple denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/litecoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicate denoms",
			genState: &types.GenesisState{
				FactoryDenoms: []types.GenesisDenom{
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
					{
						Denom: "factory/symphony1p822vyk8ylf3hpwh9qgv6p6dft7hedntyqyw7w/bitcoin",
						AuthorityMetadata: types.DenomAuthorityMetadata{
							Admin: "",
						},
					},
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
