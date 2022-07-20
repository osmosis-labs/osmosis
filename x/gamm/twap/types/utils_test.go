package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAllUniqueDenomPairs(t *testing.T) {
	tests := map[string]struct {
		denoms       []string
		wantedPairGT []string
		wantedPairLT []string
	}{
		"basic":    {[]string{"A", "B"}, []string{"B"}, []string{"A"}},
		"basicRev": {[]string{"B", "A"}, []string{"B"}, []string{"A"}},
		// AB > A
		"prefixed": {[]string{"A", "AB"}, []string{"AB"}, []string{"A"}},
		"basic-3":  {[]string{"A", "B", "C"}, []string{"C", "C", "B"}, []string{"B", "A", "A"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			pairGT, pairLT := GetAllUniqueDenomPairs(tt.denoms)
			require.Equal(t, pairGT, tt.wantedPairGT)
			require.Equal(t, pairLT, tt.wantedPairLT)
		})
	}
}
