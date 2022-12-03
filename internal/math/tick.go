package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/osmomath"
)

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
	if tickIndex.GTE(sdk.ZeroInt()) {
		sqrtPrice, err = sdk.NewDecWithPrec(10001, 4).Power(tickIndex.Uint64()).ApproxSqrt()
	} else {
		sqrtPrice, err = sdk.OneDec().Quo(sdk.NewDecWithPrec(10001, 4).Power(2)).ApproxSqrt()
	}
	if err != nil {
		return sdk.Dec{}, err
	}

	return sqrtPrice, nil
}

// PriceToTick takes a price and returns the corresponding tick index
func PriceToTick(price sdk.Dec) sdk.Int {
	tick := osmomath.BigDecFromSDKDec(price).CustomBaseLog(osmomath.NewDecWithPrec(10001, 4))
	return tick.SDKDec().TruncateInt()
}
