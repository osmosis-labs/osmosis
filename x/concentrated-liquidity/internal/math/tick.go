package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

var TickBase = osmomath.MustNewDecFromStr("1.0001")

// TicksToSqrtPrice returns the sqrt price for the lower and upper ticks.
// Returns error if fails to calculate sqrt price.
// TODO: spec and tests
func TicksToSqrtPrice(lowerTick, upperTick int64) (sdk.Dec, sdk.Dec, error) {
	sqrtPriceUpperTick, err := TickToSqrtPrice(sdk.NewInt(upperTick))
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	sqrtPriceLowerTick, err := TickToSqrtPrice(sdk.NewInt(lowerTick))
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}
	return sqrtPriceLowerTick, sqrtPriceUpperTick, nil
}

// TickToSqrtPrice takes the tick index and returns the corresponding sqrt of the price.
// Returns error if fails to calculate sqrt price. Otherwise, the computed value and nil.
// TODO: test
func TickToSqrtPrice(tickIndex sdk.Int) (sqrtPrice sdk.Dec, err error) {
	// If the tick index is positive, we can use the normal equation to calculate the square root price.
	// However, if the tick index is negative, since we cannot take negative powers with the sdk,
	// we need to take one over the original equation in order to make the power positive.
	var sqrtPriceOsmoMath osmomath.BigDec
	if tickIndex.GTE(sdk.ZeroInt()) {
		sqrtPriceOsmoMath, err = TickBase.PowerInteger(tickIndex.Uint64()).ApproxSqrt()
	} else {
		sqrtPriceOsmoMath, err = osmomath.OneDec().Quo(TickBase.PowerInteger(tickIndex.Abs().Uint64())).ApproxSqrt()
	}
	if err != nil {
		return sdk.Dec{}, err
	}

	sqrtPrice = sqrtPriceOsmoMath.SDKDec()

	return sqrtPrice, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec) sdk.Int {
	tick := osmomath.BigDecFromSDKDec(price).CustomBaseLog(TickBase)
	return tick.SDKDec().TruncateInt()
}
