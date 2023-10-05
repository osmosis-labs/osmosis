package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

func TestFormatDenomTradePairKey(t *testing.T) {
	tests := map[string]struct {
		denom0      string
		denom1      string
		expectedKey string
	}{
		"happy path": {
			denom0:      "uosmo",
			denom1:      "uion",
			expectedKey: "\x04|uion|uosmo",
		},
		"reversed denoms get reordered": {
			denom0:      "uion",
			denom1:      "uosmo",
			expectedKey: "\x04|uion|uosmo",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			formatDenomTradePairKey := types.FormatDenomTradePairKey(tc.denom0, tc.denom1)
			stringFormatDenomTradePairKeyString := string(formatDenomTradePairKey)
			require.Equal(t, tc.expectedKey, stringFormatDenomTradePairKeyString)
		})
	}
}
