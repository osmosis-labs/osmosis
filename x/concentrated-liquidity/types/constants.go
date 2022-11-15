package types

import (
	"math"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
)

// TODO: decide on the values for Max tick and Min tick
var (
	MinTick int64 = -887272
	MaxTick int64 = 887272

	MaxSqrtRatio = sdk.MustNewDecFromStr("18446050711097703529.7763428")
	// TODO: this is a temp value, figure out math for this.
	// we basically want getSqrtRatioAtTick(MIN_TICK)
	MinSqrtRatio = GetMinSqrtRatio()
)

// Calculates MinSqrtPrice = sqrt(1.0001^MinTick)
func GetMinSqrtRatio() sdk.Dec {
	minSqrtRatio := osmomath.MustNewDecFromStr(strconv.FormatFloat(math.Pow(1.0001, -887272/2), 'f', 36, 64))
	return minSqrtRatio.SDKDec()
}
