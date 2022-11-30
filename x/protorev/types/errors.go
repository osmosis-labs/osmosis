package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidRoute       = sdkerrors.Register(ModuleName, 1, "This is not a valid cyclic route")
	ErrDuplicateTokenPair = sdkerrors.Register(ModuleName, 2, "This token pair has already been added")
	ErrInvalidTokenName   = sdkerrors.Register(ModuleName, 3, "This is not a valid token name")
	ErrInvalidParams      = sdkerrors.Register(ModuleName, 4, "Invalid params")
	ErrInvalidArbDenom    = sdkerrors.Register(ModuleName, 5, "Invalid denom for arbitrage")
)
