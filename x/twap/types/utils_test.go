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
		denoms      []string
		wantedPairs []DenomPair
		panics      bool
	}{
<<<<<<< HEAD
		"basic":    {[]string{"A", "B"}, []string{"B"}, []string{"A"}, false},
		"basicRev": {[]string{"B", "A"}, []string{"B"}, []string{"A"}, false},
		// AB > A
		"prefixed": {[]string{"A", "AB"}, []string{"AB"}, []string{"A"}, false},
		"basic-3":  {[]string{"A", "B", "C"}, []string{"C", "C", "B"}, []string{"B", "A", "A"}, false},
		"panics":   {[]string{"A", "A"}, []string{}, []string{}, true},
=======
		"basic":    {[]string{"A", "B"}, []DenomPair{{"A", "B"}}, false},
		"basicRev": {[]string{"B", "A"}, []DenomPair{{"A", "B"}}, false},
		// AB > A
		"prefixed": {[]string{"A", "AB"}, []DenomPair{{"A", "AB"}}, false},
		"basic-3":  {[]string{"A", "B", "C"}, []DenomPair{{"A", "B"}, {"A", "C"}, {"B", "C"}}, false},
		"panics":   {[]string{"A", "A"}, []DenomPair{}, true},
>>>>>>> b703f471 (TWAP code improvements (#3231))
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			osmoassert.ConditionalPanic(t, tt.panics, func() {
				pairs := GetAllUniqueDenomPairs(tt.denoms)
				require.Equal(t, pairs, tt.wantedPairs)
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
		"realDenoms": {"uosmo", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", "uosmo", nil},
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
