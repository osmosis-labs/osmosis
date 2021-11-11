package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tokenfactory module sentinel errors
var (
	ErrDenomExists = sdkerrors.Register(ModuleName, 1, "denom already exists")
)
