package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/lockup module sentinel errors.
var (
	ErrNotLockOwner                      = sdkerrors.Register(ModuleName, 1, "msg sender is not the owner of specified lock")
	ErrSyntheticLockupAlreadyExists      = sdkerrors.Register(ModuleName, 2, "synthetic lockup already exists for same lock and suffix")
	ErrSyntheticDurationLongerThanNative = sdkerrors.Register(ModuleName, 3, "synthetic lockup duration should be shorter than native lockup duration")
	ErrLockupNotFound                    = sdkerrors.Register(ModuleName, 4, "lockup not found")
	ErrNewDurationNotGreater             = sdkerrors.Register(ModuleName, 5, "new duration should be greater than the original")
	ErrRequestedExceedsTokens            = sdkerrors.Register(ModuleName, 6, "requested amount to unlock exceeds locked tokens")
	ErrNotStartedUnlocking               = sdkerrors.Register(ModuleName, 7, "lock hasn't started unlocking yet")
	ErrNotUnlockableYet                  = sdkerrors.Register(ModuleName, 8, "lock is not unlockable yet")
	ErrSyntheticLockupNotExists          = sdkerrors.Register(ModuleName, 9, "synthetic lockup does not exist")
)
