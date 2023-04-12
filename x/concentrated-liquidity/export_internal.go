package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
)

func TickToSqrtPrice(tickIndex, exponentAtPriceOne sdk.Int) (price sdk.Dec, err error) {
	return math.TickToSqrtPrice(tickIndex, exponentAtPriceOne)
}
