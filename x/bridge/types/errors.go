package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidAsset       = errorsmod.Register(ModuleName, 2, "invalid asset")
	ErrInvalidDenom       = errorsmod.Register(ModuleName, 3, "invalid denom")
	ErrInvalidSourceChain = errorsmod.Register(ModuleName, 3, "invalid source chain")
)
