package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// TestAsset tests if Asset is properly validated.
func TestAsset(t *testing.T) {
	testCases := []struct {
		name          string
		asset         types.Asset
		expectedValid bool
	}{
		{
			name:          "default is valid",
			asset:         types.DefaultAssets()[0],
			expectedValid: true,
		},
		{
			name: "empty source chain",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: "",
					Denom:       types.DefaultBitcoinDenomName,
				},
				Status:   types.AssetStatus_ASSET_STATUS_OK,
				Exponent: types.DefaultBitcoinExponent,
			},
			expectedValid: false,
		},
		{
			name: "empty denom",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: types.DefaultBitcoinChainName,
					Denom:       "",
				},
				Status:   types.AssetStatus_ASSET_STATUS_OK,
				Exponent: types.DefaultBitcoinExponent,
			},
			expectedValid: false,
		},
		{
			name: "invalid status",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: types.DefaultBitcoinChainName,
					Denom:       types.DefaultBitcoinDenomName,
				},
				Status:   types.AssetStatus_ASSET_STATUS_UNSPECIFIED,
				Exponent: types.DefaultBitcoinExponent,
			},
			expectedValid: false,
		},
		{
			name: "unknown status",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: types.DefaultBitcoinChainName,
					Denom:       types.DefaultBitcoinDenomName,
				},
				Status:   999,
				Exponent: types.DefaultBitcoinExponent,
			},
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.asset.Validate()
			if tc.expectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
