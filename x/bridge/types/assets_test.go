package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

// TestAsset tests if Asset is properly validated.
func TestAsset(t *testing.T) {
	testCases := []struct {
		name        string
		asset       types.Asset
		expectedErr error
	}{
		{
			name:        "default is valid",
			asset:       types.DefaultAssets()[0],
			expectedErr: nil,
		},
		{
			name: "empty source chain",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: "",
					Denom:       types.DefaultBitcoinDenomName,
				},
				Status:                types.AssetStatus_ASSET_STATUS_OK,
				Exponent:              types.DefaultBitcoinExponent,
				ExternalConfirmations: types.DefaultBitcoinConfirmations,
			},
			expectedErr: types.ErrInvalidAssetID,
		},
		{
			name: "empty denom",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: types.DefaultBitcoinChainName,
					Denom:       "",
				},
				Status:                types.AssetStatus_ASSET_STATUS_OK,
				Exponent:              types.DefaultBitcoinExponent,
				ExternalConfirmations: types.DefaultBitcoinConfirmations,
			},
			expectedErr: types.ErrInvalidAssetID,
		},
		{
			name: "invalid status",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: types.DefaultBitcoinChainName,
					Denom:       types.DefaultBitcoinDenomName,
				},
				Status:                types.AssetStatus_ASSET_STATUS_UNSPECIFIED,
				Exponent:              types.DefaultBitcoinExponent,
				ExternalConfirmations: types.DefaultBitcoinConfirmations,
			},
			expectedErr: types.ErrInvalidAssetStatus,
		},
		{
			name: "unknown status",
			asset: types.Asset{
				Id: types.AssetID{
					SourceChain: types.DefaultBitcoinChainName,
					Denom:       types.DefaultBitcoinDenomName,
				},
				Status:                999,
				Exponent:              types.DefaultBitcoinExponent,
				ExternalConfirmations: types.DefaultBitcoinConfirmations,
			},
			expectedErr: types.ErrInvalidAssetStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.asset.Validate()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}
