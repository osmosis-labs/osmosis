package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/authenticator module sentinel errors
var (
	ErrExtensionNotFound = errorsmod.Register(ModuleName, 2, "authenticator tx extension not found")
)
