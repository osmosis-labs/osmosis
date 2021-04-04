package types

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestGetPoolShareDenom(t *testing.T) {
	denom := GetPoolShareDenom(0)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/0", denom)

	denom = GetPoolShareDenom(10)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/10", denom)

	denom = GetPoolShareDenom(math.MaxUint64)
	require.NoError(t, sdk.ValidateDenom(denom))
	require.Equal(t, "gamm/pool/18446744073709551615", denom)
}

func TestGetPoolIdFromShareDenom(t *testing.T) {
	denom := "gamm/pool/1"

	poolId, err := GetPoolIdFromShareDenom(denom)
	require.NoError(t, err)
	require.Equal(t, uint64(1), poolId)

	_, err = GetPoolIdFromShareDenom("hello")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool/")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool/hello")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool//")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool//1")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool/1/1")
	require.Error(t, err)

	_, err = GetPoolIdFromShareDenom("gamm/pool/1/hello")
	require.Error(t, err)
}
