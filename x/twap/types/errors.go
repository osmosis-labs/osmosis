package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/twap module sentinel errors.
var (
	ErrKeySeparatorLength  = sdkerrors.Register(ModuleName, 2, "key separator is an incorrect length")
	ErrUnexpectedSeparator = sdkerrors.Register(ModuleName, 3, "separator is incorrectly formatted")
)
