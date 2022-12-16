package osmomath

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	EulersNumber = eulersNumber
	TwoBigDec    = twoBigDec
)

// 2^128 - 1, needs to be the same as gammtypes.MaxSpotPrice
// but we can't directly import that due to import cycles.
// Hence we use the same var name, in hopes that if any change there happens,
// this is caught via a CTRL+F
var MaxSpotPrice = sdk.NewDec(2).Power(128).Sub(sdk.OneDec())
