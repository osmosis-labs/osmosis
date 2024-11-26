package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
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
		"basic":    {[]string{"A", "B"}, []DenomPair{{"A", "B"}}, false},
		"basicRev": {[]string{"B", "A"}, []DenomPair{{"A", "B"}}, false},
		// AB > A
		"prefixed": {[]string{"A", "AB"}, []DenomPair{{"A", "AB"}}, false},
		"basic-3":  {[]string{"A", "B", "C"}, []DenomPair{{"A", "B"}, {"A", "C"}, {"B", "C"}}, false},
		"panics":   {[]string{"A", "A"}, []DenomPair{}, true},
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

func TestCanonicalTimeMs(t *testing.T) {
	const expectedMs int64 = 2

	newYorkLocation, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	time := time.Unix(0, int64(time.Millisecond+999999+1)).In(newYorkLocation)

	actualTime := CanonicalTimeMs(time)
	require.Equal(t, expectedMs, actualTime)
}
