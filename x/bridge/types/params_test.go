package types_test

import (
	"testing"

	math "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v24/x/bridge/types"
)

// TestParams tests if Params are properly validated.
func TestParams(t *testing.T) {
	var testCases = []struct {
		name        string
		params      types.Params
		expectedErr error
	}{
		{
			name: "valid",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{asset1, asset2},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(5, 1),
			},
			expectedErr: nil,
		},
		{
			name: "duplicated signers",
			params: types.Params{
				Signers:     []string{addr1, addr1},
				Assets:      []types.Asset{asset1, asset2},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(5, 1),
			},
			expectedErr: types.ErrInvalidSigners,
		},
		{
			name: "empty assets",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(5, 1),
			},
			expectedErr: types.ErrInvalidAssets,
		},
		{
			name: "invalid asset",
			params: types.Params{
				Signers: []string{addr1},
				Assets: []types.Asset{{
					Id:       assetID1,
					Status:   types.AssetStatus_ASSET_STATUS_UNSPECIFIED, // invalid status
					Exponent: types.DefaultBitcoinExponent,
				}},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(5, 1),
			},
			expectedErr: types.ErrInvalidAssets,
		},
		{
			name: "duplicated assets",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{asset1, asset1},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(5, 1),
			},
			expectedErr: types.ErrInvalidAssets,
		},
		{
			name: "fee == 0 valid",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{asset1},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyZeroDec(),
			},
			expectedErr: nil,
		},
		{
			name: "fee == 1 valid",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{asset1},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyOneDec(),
			},
			expectedErr: nil,
		},
		{
			name: "fee < 0",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{asset1},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(-5, 1),
			},
			expectedErr: types.ErrInvalidFee,
		},
		{
			name: "fee > 1",
			params: types.Params{
				Signers:     []string{addr1},
				Assets:      []types.Asset{asset1},
				VotesNeeded: types.DefaultVotesNeeded,
				Fee:         math.LegacyNewDecWithPrec(11, 1),
			},
			expectedErr: types.ErrInvalidFee,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			require.ErrorIsf(t, err, tc.expectedErr, "test: %v", tc.name)
		})
	}
}
