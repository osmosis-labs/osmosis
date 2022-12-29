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
	swapRouterExactAmountInRoute := swaproutertypes.SwapAmountInRoute{}
	gammExactAmountInRoute := gammtypes.SwapAmountInRoute{}

	swapRouterBz, err := proto.Marshal(&swapRouterExactAmountInRoute)
	require.NoError(t, err)
	gammBz, err := proto.Marshal(&gammExactAmountInRoute)
	require.NoError(t, err)

	require.Equal(t, swapRouterBz, gammBz)
}
