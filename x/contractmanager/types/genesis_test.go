package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v20/x/contractmanager/types"
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
				FailuresList: []types.Failure{
					{
						Address: "address1",
						Id:      1,
					},
					{
						Address: "address1",
						Id:      2,
					},
					{
						Address: "address2",
						Id:      1,
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicated failure",
			genState: &types.GenesisState{
				FailuresList: []types.Failure{
					{
						Address: "address1",
						Id:      1,
					},
					{
						Address: "address1",
						Id:      1,
					},
					{
						Address: "address2",
						Id:      1,
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
