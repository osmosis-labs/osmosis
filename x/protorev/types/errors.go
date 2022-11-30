package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidArbDenom     = sdkerrors.Register(ModuleName, 1, "This is not a tradeable denomination")
	ErrInvalidRoute        = sdkerrors.Register(ModuleName, 2, "This is not a valid cyclic route")
	ErrDuplicateTokenPair  = sdkerrors.Register(ModuleName, 3, "This token pair has already been added")
	ErrInvalidTokenName    = sdkerrors.Register(ModuleName, 4, "This is not a valid token name")
	ErrInvalidParams       = sdkerrors.Register(ModuleName, 5, "Invalid params")
	ErrInvalidArbRoutes    = sdkerrors.Register(ModuleName, 6, "Invalid routes")
	ErrInvalidBlockHeights = sdkerrors.Register(ModuleName, 7, "Invalid block heights")
)
