package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/txfees module errors.
var (
	ErrNoBaseDenom     = sdkerrors.Register(ModuleName, 1, "no base denom was set")
	ErrTooManyFeeCoins = sdkerrors.Register(ModuleName, 2, "too many fee coins. only accepts fees in one denom")
	ErrInvalidFeeToken = sdkerrors.Register(ModuleName, 3, "invalid fee token")
)
