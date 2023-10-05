package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DONTCOVER

// x/txfees module errors.
var (
	ErrNoBaseDenom     = errorsmod.Register(ModuleName, 1, "no base denom was set")
	ErrTooManyFeeCoins = errorsmod.Register(ModuleName, 2, "too many fee coins. only accepts fees in one denom")
	ErrInvalidFeeToken = errorsmod.Register(ModuleName, 3, "invalid fee token")
)
