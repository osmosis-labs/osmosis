package types_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// TestParams tests if Params are properly validated.
func TestParams(t *testing.T) {
	var (
		pk1        = ed25519.GenPrivKey().PubKey()
		addr1Bytes = sdk.AccAddress(pk1.Address())
		addr1      = addr1Bytes.String()

		assetID1 = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "wbtc1",
		}
		assetID2 = types.AssetID{
			SourceChain: "bitcoin",
			Denom:       "wbtc2",
		}
	)
	asset1 := types.DefaultAssets()[0]
	asset1.Id = assetID1

	asset2 := types.DefaultAssets()[0]
	asset2.Id = assetID2

	// don't check the invalid asset cases here since it is already tested in a different case
	var testCases = []struct {
		name          string
		params        types.Params
		expectedValid bool
	}{
		{
			name: "valid",
			params: types.Params{
				Signers: []string{addr1},
				Assets:  []types.Asset{asset1, asset2},
			},
			expectedValid: true,
		},
		{
			name: "duplicated signers",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets:  []types.Asset{asset1, asset2},
			},
			expectedValid: false,
		},
		{
			name: "empty assets",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets:  []types.Asset{},
			},
			expectedValid: false,
		},
		{
			name: "invalid asset",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets: []types.Asset{{
					Id:          assetID1,
					Status:      types.AssetStatus_ASSET_STATUS_OK,
					Precision:   10,
					VotesNeeded: 1,
					Fee:         200, // fee > 100
				}},
			},
			expectedValid: false,
		},
		{
			name: "duplicated assets",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets:  []types.Asset{asset1, asset1},
			},
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedValid {
				require.NoError(t, tc.params.Validate(), "test: %v", tc.name)
			} else {
				require.Error(t, tc.params.Validate(), "test: %v", tc.name)
			}
		})
	}
}
