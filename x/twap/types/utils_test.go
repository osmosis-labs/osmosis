package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func TestMaxSpotPriceEquality(t *testing.T) {
	require.Equal(t, MaxSpotPrice, types.MaxSpotPrice)
}

func TestGetAllUniqueDenomPairs(t *testing.T) {
	tests := map[string]struct {
		denoms       []string
		wantedPairGT []string
		wantedPairLT []string
		panics       bool
	}{
		"basic":    {[]string{"A", "B"}, []string{"A"}, []string{"B"}, false},
		"basicRev": {[]string{"B", "A"}, []string{"A"}, []string{"B"}, false},
		// AB > A
		"prefixed": {[]string{"A", "AB"}, []string{"A"}, []string{"AB"}, false},
		"basic-3":  {[]string{"A", "B", "C"}, []string{"A", "A", "B"}, []string{"B", "C", "C"}, false},
		"panics":   {[]string{"A", "A"}, []string{}, []string{}, true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			osmoassert.ConditionalPanic(t, tt.panics, func() {
				pairGT, pairLT := GetAllUniqueDenomPairs(tt.denoms)
				require.Equal(t, pairGT, tt.wantedPairGT)
				require.Equal(t, pairLT, tt.wantedPairLT)
			})
		})
	}
}

func TestLexicographicalOrderDenoms(t *testing.T) {
	tests := map[string]struct {
		firstDenom     string
		secondDenom    string
		expectedDenomA string
		expectedDenomB string
		expectedErr    error
	}{
		"basic":    {"A", "B", "A", "B", nil},
		"basicRev": {"B", "A", "A", "B", nil},
		"realDenoms": {
			"uosmo", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", "uosmo", nil,
		},
		"sameDenom": {"A", "A", "", "", fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", "A", "A")},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// system under test
			denomA, denomB, err := LexicographicalOrderDenoms(tt.firstDenom, tt.secondDenom)
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.Equal(t, denomA, tt.expectedDenomA)
				require.Equal(t, denomB, tt.expectedDenomB)
			}
		})
	}
}
