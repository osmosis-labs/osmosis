package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

func TestFormatDenomTradePairKey(t *testing.T) {
	tests := map[string]struct {
		tokenInDenom  string
		tokenOutDenom string
		expectedKey   string
	}{
		"happy path": {
			tokenInDenom:  appparams.BaseCoinUnit,
			tokenOutDenom: "uion",
			expectedKey:   "\x04|uosmo|uion",
		},
		"reversed denoms no longer get reordered": {
			tokenInDenom:  "uion",
			tokenOutDenom: appparams.BaseCoinUnit,
			expectedKey:   "\x04|uion|uosmo",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			formatDenomTradePairKey := types.FormatDenomTradePairKey(tc.tokenInDenom, tc.tokenOutDenom)
			stringFormatDenomTradePairKeyString := string(formatDenomTradePairKey)
			require.Equal(t, tc.expectedKey, stringFormatDenomTradePairKeyString)
		})
	}
}

func TestParseDenomTradePairKey(t *testing.T) {
	// Define a valid DenomTradePairKey
	key := fmt.Sprintf("%s%s%s%s%s", types.DenomTradePairPrefix, types.KeySeparator, "tokenInDenom", types.KeySeparator, "tokenOutDenom")

	// Call the function with the valid key
	tokenInDenom, tokenOutDenom, err := types.ParseDenomTradePairKey([]byte(key))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// Check the results
	if tokenInDenom != "tokenInDenom" {
		t.Errorf("Expected tokenInDenom, got %s", tokenInDenom)
	}

	if tokenOutDenom != "tokenOutDenom" {
		t.Errorf("Expected tokenOutDenom, got %s", tokenOutDenom)
	}

	// Define an invalid DenomTradePairKey
	invalidKey := fmt.Sprintf("%s%s%s%s%s", types.DenomTradePairPrefix, types.KeySeparator, "denom0!_", types.KeySeparator, "denom1!_")

	_, _, err = types.ParseDenomTradePairKey([]byte(invalidKey))
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestFormatModuleRouteKey(t *testing.T) {
	cases := []struct {
		id                 uint64
		expectedSansPrefix string
	}{0: {id: 0, expectedSansPrefix: "0"},
		1: {id: 1, expectedSansPrefix: "1"},
		2: {id: 12, expectedSansPrefix: "12"},
		3: {id: 122, expectedSansPrefix: "122"},
		4: {id: 4522, expectedSansPrefix: "4522"},
		5: {id: 54522, expectedSansPrefix: "54522"},
		6: {id: 654522, expectedSansPrefix: "654522"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("id=%d", tc.id), func(t *testing.T) {
			key := types.FormatModuleRouteKey(tc.id)
			require.Equal(t, types.SwapModuleRouterPrefix[0], key[0])
			require.Equal(t, tc.expectedSansPrefix, string(key[1:]))
		})
	}
}
