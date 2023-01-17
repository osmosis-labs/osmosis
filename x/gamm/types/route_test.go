package types_test

import (
	"encoding/json"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

// TestSwapRoutesSerialization tests that while swap routes
// are proto-generated from different modules, they are identical in terms
// of serialization and deserialization.
func TestSwapRoutes_MarshalUnmarshal(t *testing.T) {
	const (
		testPoolId        = 2
		testTokenOutDenom = "uosmo"
	)
	poolManagertypesExactAmountInRoute := poolmanagertypes.SwapAmountInRoute{
		PoolId:        testPoolId,
		TokenOutDenom: testTokenOutDenom,
	}
	gammExactAmountInRoute := poolmanagertypes.SwapAmountInRoute{
		PoolId:        testPoolId,
		TokenOutDenom: testTokenOutDenom,
	}

	poolManagerBz, err := proto.Marshal(&poolManagertypesExactAmountInRoute)
	require.NoError(t, err)
	gammBz, err := proto.Marshal(&gammExactAmountInRoute)
	require.NoError(t, err)

	require.Equal(t, poolManagerBz, gammBz)

	jsonPoolmanagerBz, err := json.Marshal(&poolManagertypesExactAmountInRoute)
	require.NoError(t, err)

	jsonGammBz, err := json.Marshal(&gammExactAmountInRoute)
	require.NoError(t, err)
	require.Equal(t, jsonPoolmanagerBz, jsonGammBz)
}
