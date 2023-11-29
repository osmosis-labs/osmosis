package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
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

func TestParseDenomTradePairKey(t *testing.T) {
	// Define a valid DenomTradePairKey
	key := fmt.Sprintf("%s%s%s%s%s", types.DenomTradePairPrefix, types.KeySeparator, "denom0", types.KeySeparator, "denom1")

	// Call the function with the valid key
	denom0, denom1, err := types.ParseDenomTradePairKey([]byte(key))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// Check the results
	if denom0 != "denom0" {
		t.Errorf("Expected denom0, got %s", denom0)
	}

	if denom1 != "denom1" {
		t.Errorf("Expected denom1, got %s", denom1)
	}

	// Define an invalid DenomTradePairKey
	invalidKey := fmt.Sprintf("%s%s%s%s%s", types.DenomTradePairPrefix, types.KeySeparator, "denom0!_", types.KeySeparator, "denom1!_")

	_, _, err = types.ParseDenomTradePairKey([]byte(invalidKey))
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
