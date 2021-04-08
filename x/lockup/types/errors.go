package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/lockup module sentinel errors
var (
	ErrSample = sdkerrors.Register(ModuleName, 1100, "sample error")

	ErrNotLockOwner = sdkerrors.Register(ModuleName, 1, "msg sender is not the owner of specified lock")
)
