package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidSubdenom = errorsmod.Register(ModuleName, 2, "invalid subdenom")
	ErrInvalidChain    = errorsmod.Register(ModuleName, 3, "invalid chain")
)
