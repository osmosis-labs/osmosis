package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: decide on the values for Max tick and Min tick
var (
	MinTick int64 = -887272
	MaxTick int64 = 887272

	MaxSqrtRatio = sdk.MustNewDecFromStr("18446050711097703529.7763428")
	// TODO: this is a temp value, figure out math for this.
	// we basically want getSqrtRatioAtTick(MIN_TICK)
	MinSqrtRatio              = sdk.MustNewDecFromStr("0")
	ConcentratedGasFeeForSwap = 10_000
	UpperPriceLimit           = sdk.NewDec(999999999999)
	LowerPriceLimit           = sdk.NewDec(1)
)
