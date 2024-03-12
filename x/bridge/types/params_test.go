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

	// don't check the invalid asset case here since it is already tested in a different case
	var testCases = []struct {
		name          string
		params        types.Params
		expectedValid bool
	}{
		{
			name: "expectedValid",
			params: types.Params{
				Signers: []string{addr1},
				Assets: []types.AssetWithStatus{
					{asset1, types.AssetStatus_ASSET_STATUS_OK},
					{asset2, types.AssetStatus_ASSET_STATUS_OK},
				},
			},
			expectedValid: true,
		},
		{
			name: "duplicated signers",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets: []types.AssetWithStatus{
					{asset1, types.AssetStatus_ASSET_STATUS_OK},
					{asset2, types.AssetStatus_ASSET_STATUS_OK},
				},
			},
			expectedValid: false,
		},
		{
			name: "empty assets",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets:  []types.AssetWithStatus{},
			},
			expectedValid: false,
		},
		{
			name: "duplicated assets",
			params: types.Params{
				Signers: []string{addr1, addr1},
				Assets: []types.AssetWithStatus{
					{asset1, types.AssetStatus_ASSET_STATUS_OK},
					{asset1, types.AssetStatus_ASSET_STATUS_OK},
				},
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
