package types_test

import (
	"testing"

	math "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

// TestGenesisState tests if GenesisState is properly validated.
func TestGenesisState(t *testing.T) {
	testCases := []struct {
		name        string
		genState    types.GenesisState
		expectedErr error
	}{
		{
			name:        "default is valid",
			genState:    *types.DefaultGenesis(),
			expectedErr: nil,
		},
		{
			name: "duplicated signers",
			genState: types.GenesisState{
				Params: types.Params{
					Signers:     []string{addr1, addr1},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedErr: types.ErrInvalidSigners,
		},
		{
			name: "empty assets",
			genState: types.GenesisState{
				Params: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedErr: types.ErrInvalidAssets,
		},
		{
			name: "invalid asset",
			genState: types.GenesisState{
				Params: types.Params{
					Signers: []string{addr1},
					Assets: []types.Asset{{
						Id:       assetID1,
						Status:   types.AssetStatus_ASSET_STATUS_UNSPECIFIED, // invalid status
						Exponent: types.DefaultBitcoinExponent,
					}},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedErr: types.ErrInvalidAssets,
		},
		{
			name: "duplicated assets",
			genState: types.GenesisState{
				Params: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{asset1, asset1},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedErr: types.ErrInvalidAssets,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genState.Validate()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}
