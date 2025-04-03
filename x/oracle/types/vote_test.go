package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

func TestParseExchangeRateTuples(t *testing.T) {
	valid := "123.0unote,123.123ukrw"
	_, err := types.ParseExchangeRateTuples(valid)
	require.NoError(t, err)

	duplicatedDenom := "100.0unote,123.123ukrw,121233.123ukrw"
	_, err = types.ParseExchangeRateTuples(duplicatedDenom)
	require.Error(t, err)

	invalidCoins := "123.123"
	_, err = types.ParseExchangeRateTuples(invalidCoins)
	require.Error(t, err)

	invalidCoinsWithValid := "123.0unote,123.1"
	_, err = types.ParseExchangeRateTuples(invalidCoinsWithValid)
	require.Error(t, err)

	abstainCoinsWithValid := "0.0unote,123.1ukrw"
	_, err = types.ParseExchangeRateTuples(abstainCoinsWithValid)
	require.NoError(t, err)
}
