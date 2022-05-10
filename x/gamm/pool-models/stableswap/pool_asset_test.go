package stableswap_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/stableswap"
)

func TestStableSwapPoolAssetValidate(t *testing.T) {
	testcase := map[string]struct {
		name          string
		tokenAmount   int64
		scalingFactor int64
		expected      error
	}{
		"valid pool asset": {
			name:          "ust",
			tokenAmount:   100,
			scalingFactor: 10,
		},
		"zero scaling factor - invalid": {
			name:          "ust",
			tokenAmount:   100,
			scalingFactor: 0,
			expected:      fmt.Errorf(stableswap.ErrMsgFmtNonPositiveScalingFactor, "ust", 0),
		},
		"zero token amount - invalid": {
			name:          "ust",
			tokenAmount:   0,
			scalingFactor: 10,
			expected:      fmt.Errorf(stableswap.ErrMsgFmtNonPositiveTokenAmount, "ust", 0),
		},
	}
	for name, tc := range testcase {
		t.Run(name, func(t *testing.T) {
			poolAsset := stableswap.PoolAsset{
				Token:         sdk.NewCoin(tc.name, sdk.NewInt(tc.tokenAmount)),
				ScalingFactor: sdk.NewInt(tc.scalingFactor),
			}

			res := poolAsset.Validate()

			require.Equal(t, tc.expected, res)
		})
	}
}
