package stableswap_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
	types "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func TestStableswapValidatePoolAssets(t *testing.T) {
	testcases := map[string]struct {
		poolAssets    []stableswap.PoolAsset
		expectedError error
	}{
		"valid pool assets - success": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
		},
		"one pool asset - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
			},
			expectedError: types.ErrTooFewPoolAssets,
		},
		"zero pool assets - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
			},
			expectedError: types.ErrTooFewPoolAssets,
		},
		"three pool assets - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("usdt", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
			expectedError: types.ErrTooManyPoolAssets,
		},
		"empty denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.Coin{Denom: "", Amount: sdk.NewInt(100000000)},
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
			expectedError: types.ErrEmptyPoolAssets,
		},
		"duplicate denom - error": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100000000)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtDuplicateDenomFound, "usdc"),
		},
		"non-positive token amount": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(0)),
					ScalingFactor: sdk.NewInt(100000),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveTokenAmount, "usdc", 0),
		},
		"non-positive scaling factor": {
			poolAssets: []stableswap.PoolAsset{
				{
					Token:         sdk.NewCoin("usdc", sdk.NewInt(1000)),
					ScalingFactor: sdk.NewInt(0),
				},
				{
					Token:         sdk.NewCoin("ust", sdk.NewInt(100)),
					ScalingFactor: sdk.NewInt(1),
				},
			},
			expectedError: fmt.Errorf(stableswap.ErrMsgFmtNonPositiveScalingFactor, "usdc", 0),
		},
	}

	for _, tc := range testcases {
		err := stableswap.ValidatePoolAssets(tc.poolAssets)
		require.Equal(t, tc.expectedError, err)
	}

}
