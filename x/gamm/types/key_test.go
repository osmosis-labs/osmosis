package types_test

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

func TestGetPoolShareDenom(t *testing.T) {
	denom := types.GetPoolShareDenom(0)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/0", denom)

	denom = types.GetPoolShareDenom(10)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/10", denom)

	denom = types.GetPoolShareDenom(math.MaxUint64)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/18446744073709551615", denom)
}

func TestGetPoolIdFromShareDenom(t *testing.T) {
	tests := []struct {
		name          string
		denom         string
		expectedId    uint64
		expectedError string
	}{
		{
			name:       "Valid pool id",
			denom:      "gamm/pool/123",
			expectedId: 123,
		},
		{
			name:          "Invalid pool id",
			denom:         "gamm/pool/abc",
			expectedId:    0,
			expectedError: "strconv.Atoi: parsing \"bc\": invalid syntax",
		},
		{
			name:          "Empty pool id",
			denom:         "gamm/pool/",
			expectedId:    0,
			expectedError: "strconv.Atoi: parsing \"\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := types.GetPoolIdFromShareDenom(tt.denom)
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			if id != tt.expectedId || errStr != tt.expectedError {
				t.Errorf("GetPoolIdFromShareDenom(%q) = (%v, %q), want (%v, %q)",
					tt.denom, id, errStr, tt.expectedId, tt.expectedError)
			}
		})
	}
}
