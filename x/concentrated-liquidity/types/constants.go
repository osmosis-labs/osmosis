package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	MaxSqrtRatio = sdk.MustNewDecFromStr("18446050711097703529.7763428")
	// TODO: this is a temp value, figure out math for this.
	// we basically want getSqrtRatioAtTick(MIN_TICK)
	MinSqrtRatio              = sdk.MustNewDecFromStr("0")
	ConcentratedGasFeeForSwap = 10_000
	UpperPriceLimit           = sdk.NewDec(999999999999)
	LowerPriceLimit           = sdk.NewDec(1)
	ExponentAtPriceOneMax     = sdk.NewInt(-1)
	ExponentAtPriceOneMin     = sdk.NewInt(-12)
	MaxSpotPrice              = sdk.MustNewDecFromStr("100000000000000000000000000000000000000")
	MinSpotPrice              = sdk.MustNewDecFromStr("0.000000000000000001")
)
