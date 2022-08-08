package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v10/app/apptesting/osmoassert"
)

func TestGetAllUniqueDenomPairs(t *testing.T) {
	tests := map[string]struct {
		denoms       []string
		wantedPairGT []string
		wantedPairLT []string
		panics       bool
	}{
		"basic":    {[]string{"A", "B"}, []string{"B"}, []string{"A"}, false},
		"basicRev": {[]string{"B", "A"}, []string{"B"}, []string{"A"}, false},
		// AB > A
		"prefixed": {[]string{"A", "AB"}, []string{"AB"}, []string{"A"}, false},
		"basic-3":  {[]string{"A", "B", "C"}, []string{"C", "C", "B"}, []string{"B", "A", "A"}, false},
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
