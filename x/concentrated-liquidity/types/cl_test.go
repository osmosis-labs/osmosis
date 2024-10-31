package types_test

import (
	"fmt"
	"testing"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

func TestGetConcentratedLockupDenomFromPoolId(t *testing.T) {
	testCases := []struct {
		name          string
		poolId        uint64
		expectedDenom string
	}{
		{
			name:          "poolId 1",
			poolId:        1,
			expectedDenom: fmt.Sprintf("%s/%d", types.ConcentratedLiquidityTokenPrefix, 1),
		},
		{
			name:          "poolId 0",
			poolId:        0,
			expectedDenom: fmt.Sprintf("%s/%d", types.ConcentratedLiquidityTokenPrefix, 0),
		},
		{
			name:          "poolId 1000",
			poolId:        1000,
			expectedDenom: fmt.Sprintf("%s/%d", types.ConcentratedLiquidityTokenPrefix, 1000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			denom := types.GetConcentratedLockupDenomFromPoolId(tc.poolId)
			if denom != tc.expectedDenom {
				t.Errorf("unexpected denom; got %s, want %s", denom, tc.expectedDenom)
			}
		})
	}
}
