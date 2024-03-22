package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Market errors
var (
	ErrRecursiveSwap    = sdkerrors.Register(ModuleName, 2, "recursive swap")
	ErrNoEffectivePrice = sdkerrors.Register(ModuleName, 3, "no price registered with oracle")
	ErrZeroSwapCoin     = sdkerrors.Register(ModuleName, 4, "zero swap coin")
)
