package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidArbDenom = sdkerrors.Register(ModuleName, 1, "This is not a tradeable denomination")
	ErrInvalidRoute    = sdkerrors.Register(ModuleName, 2, "This is not a valid cyclic route")
)
