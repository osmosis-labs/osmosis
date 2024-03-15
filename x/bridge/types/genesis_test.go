package types_test

import (
	"testing"

	math "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// TestGenesisState tests if GenesisState is properly validated.
func TestGenesisState(t *testing.T) {
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		assetID1 = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc1",
		}
		assetID2 = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "btc2",
		}
	)
	asset1 := types.DefaultAssets()[0]
	asset1.Id = assetID1

	asset2 := types.DefaultAssets()[0]
	asset2.Id = assetID2

	testCases := []struct {
		name          string
		genState      *types.GenesisState
		expectedValid bool
	}{
		{
			name:          "default is valid",
			genState:      types.DefaultGenesis(),
			expectedValid: true,
		},
		{
			name: "duplicated signers",
			genState: &types.GenesisState{
				Params: types.Params{
					Signers:     []string{addr1, addr1},
					Assets:      []types.Asset{asset1, asset2},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedValid: false,
		},
		{
			name: "empty assets",
			genState: &types.GenesisState{
				Params: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedValid: false,
		},
		{
			name: "invalid asset",
			genState: &types.GenesisState{
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
			expectedValid: false,
		},
		{
			name: "duplicated assets",
			genState: &types.GenesisState{
				Params: types.Params{
					Signers:     []string{addr1},
					Assets:      []types.Asset{asset1, asset1},
					VotesNeeded: types.DefaultVotesNeeded,
					Fee:         math.LegacyNewDecWithPrec(5, 1),
				},
			},
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.expectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
