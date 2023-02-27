package types_test

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

func TestGetPoolShareDenom(t *testing.T) {
	denom := types.GetPoolShareDenom(0)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/0", denom)

	denom = types.GetPoolShareDenom(10)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/10", denom)

	denom = types.GetPoolShareDenom(math.MaxUint64)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/18446744073709551615", denom)
}
