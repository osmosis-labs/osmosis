package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/epochs module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 69420, "sample error")
)
