package cliq

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/cliq/math"
)

func TickToSqrtPrice(tickIndex sdk.Int) (price sdk.Dec, err error) {
	return math.TickToSqrtPrice(tickIndex)
}
