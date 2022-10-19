package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: decide on the values for Max tick and Min tick
var (
	MaxTick = sdk.NewIntFromUint64(887272)
	MinTick = MaxTick.Neg()
)
