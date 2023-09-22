package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/authenticator module sentinel errors
var (
	ErrExtensionNotFound = sdkerrors.Register(ModuleName, 2, "authenticator tx extension not found")
)
