package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	MaxTick = sdk.NewIntFromUint64(887272)
	MinTick = MaxTick.Neg()
)
