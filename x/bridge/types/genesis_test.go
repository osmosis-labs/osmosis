package types_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func TestGenesisState_Validate(t *testing.T) {
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		asset1 = types.Asset{
			SourceChain: "bitcoin",
			Denom:       "wbtc1",
			Precision:   10,
		}
		asset2 = types.Asset{
			SourceChain: "bitcoin",
			Denom:       "wbtc2",
			Precision:   10,
		}
	)

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
			name: "expectedValid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{},
			},
			expectedValid: true,
		},
		{
			name: "duplicated signers",
			genState: &types.GenesisState{
				types.Params{
					Signers: []string{addr1, addr1},
					Assets: []types.AssetWithStatus{
						{asset1, types.AssetStatus_ASSET_STATUS_OK},
						{asset2, types.AssetStatus_ASSET_STATUS_OK},
					},
				},
			},
			expectedValid: false,
		},
		{
			name: "empty assets",
			genState: &types.GenesisState{
				types.Params{
					Signers: []string{addr1, addr1},
					Assets:  []types.AssetWithStatus{},
				},
			},
			expectedValid: false,
		},
		{
			name: "duplicated assets",
			genState: &types.GenesisState{
				types.Params{
					Signers: []string{addr1, addr1},
					Assets: []types.AssetWithStatus{
						{asset1, types.AssetStatus_ASSET_STATUS_OK},
						{asset1, types.AssetStatus_ASSET_STATUS_OK},
					},
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
