package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidAsset       = errorsmod.Register(ModuleName, 2, "invalid asset")
	ErrInvalidParams      = errorsmod.Register(ModuleName, 3, "invalid params")
	ErrInvalidDenom       = errorsmod.Register(ModuleName, 4, "invalid denom")
	ErrInvalidSourceChain = errorsmod.Register(ModuleName, 5, "invalid source chain")
)
