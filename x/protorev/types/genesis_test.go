package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

func TestGenesisStateValidate(t *testing.T) {
	validStepSize := sdk.NewInt(1_000_000)
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
				TokenPairs: []types.TokenPairArbRoutes{
					{
						ArbRoutes: []*types.Route{{
							Trades: []*types.Trade{
								{
									Pool:     1,
									TokenIn:  types.AtomDenomination,
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: types.AtomDenomination,
								},
							},
						}},
						TokenIn:  types.OsmosisDenomination,
						TokenOut: "Juno",
						StepSize: &validStepSize,
					},
				},
			},
			valid: true,
		},
		{
			description: "Default parameters with invalid routes (duplicate token pairs)",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{
					{
						ArbRoutes: []*types.Route{
							{
								Trades: []*types.Trade{
									{
										Pool:     1,
										TokenIn:  types.AtomDenomination,
										TokenOut: "Juno",
									},
									{
										Pool:     0,
										TokenIn:  "Juno",
										TokenOut: types.OsmosisDenomination,
									},
									{
										Pool:     3,
										TokenIn:  types.OsmosisDenomination,
										TokenOut: types.AtomDenomination,
									},
								},
							},
						},
						TokenIn:  types.OsmosisDenomination,
						TokenOut: "Juno",
						StepSize: &validStepSize,
					},
					{
						ArbRoutes: []*types.Route{
							{
								Trades: []*types.Trade{
									{
										Pool:     1,
										TokenIn:  types.AtomDenomination,
										TokenOut: "Juno",
									},
									{
										Pool:     0,
										TokenIn:  "Juno",
										TokenOut: types.OsmosisDenomination,
									},
									{
										Pool:     3,
										TokenIn:  types.OsmosisDenomination,
										TokenOut: types.AtomDenomination,
									},
								},
							},
						},
						TokenIn:  types.OsmosisDenomination,
						TokenOut: "Juno",
						StepSize: &validStepSize,
					},
				},
			},
			valid: false,
		},
		{
			description: "Default parameters with nil routes",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{
					{
						ArbRoutes: nil,
						TokenIn:   types.OsmosisDenomination,
						TokenOut:  "Juno",
						StepSize:  &validStepSize,
					},
				},
			},
			valid: false,
		},
		{
			description: "Default parameters with invalid routes (too few trades in a route)",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{
					{
						ArbRoutes: []*types.Route{
							{
								Trades: []*types.Trade{
									{
										Pool:     3,
										TokenIn:  types.OsmosisDenomination,
										TokenOut: types.AtomDenomination,
									},
								},
							},
						},
						TokenIn:  types.OsmosisDenomination,
						TokenOut: "Juno",
						StepSize: &validStepSize,
					},
				},
			},
			valid: false,
		},
		{
			description: "Default parameters with invalid routes (mismatch in and out denoms)",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPairArbRoutes{
					{
						ArbRoutes: []*types.Route{{
							Trades: []*types.Trade{
								{
									Pool:     1,
									TokenIn:  types.AtomDenomination,
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "eth",
								},
							},
						}},
						TokenIn:  types.OsmosisDenomination,
						TokenOut: "Juno",
						StepSize: &validStepSize,
					},
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
