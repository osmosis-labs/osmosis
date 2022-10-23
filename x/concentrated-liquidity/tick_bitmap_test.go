package concentrated_liquidity_test

import (
	"testing"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func TestTickBitmap_FlipTick(t *testing.T) {
	bitmap := cl.NewTickBitmap()

	bitmap.FlipTick(10, 2)

}

func TestTickBitmap_NextInitializedTickWithinOneWord(t *testing.T) {

}
