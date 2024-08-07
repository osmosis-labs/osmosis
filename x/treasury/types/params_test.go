package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, params.Validate())

	params = DefaultParams()
	params.WindowLong = 0
	require.Error(t, params.Validate())

	require.NotNil(t, params.ParamSetPairs())
	require.NotNil(t, params.String())
}
