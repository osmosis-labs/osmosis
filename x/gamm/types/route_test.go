package types_test

import (
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// TestSwapRoutesSerialization tests that while swap routes
// are proto-generated from different modules, they are identical in terms
// of serialization and deserialization.
func TestSwapRoutes_MarshalUnmarshal(t *testing.T) {
	const (
		testPoolId        = 2
		testTokenOutDenom = "uosmo"
	)
	swapRouterExactAmountInRoute := swaproutertypes.SwapAmountInRoute{
		PoolId:        testPoolId,
		TokenOutDenom: testTokenOutDenom,
	}
	gammExactAmountInRoute := gammtypes.SwapAmountInRoute{
		PoolId:        testPoolId,
		TokenOutDenom: testTokenOutDenom,
	}

	swapRouterBz, err := proto.Marshal(&swapRouterExactAmountInRoute)
	require.NoError(t, err)
	gammBz, err := proto.Marshal(&gammExactAmountInRoute)
	require.NoError(t, err)

	require.Equal(t, swapRouterBz, gammBz)
}
