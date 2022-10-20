package concentrated_liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func TestTickBitmap_FlipTick(t *testing.T) {
	tb := cl.NewTickBitmap()

	require.NoError(t, tb.FlipTick(85176, 1))
	require.Error(t, tb.FlipTick(85176, 1000))
}

func TestTickBitmap_NextInitializedTickWithinOneWord(t *testing.T) {

}
