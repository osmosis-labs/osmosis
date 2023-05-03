package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DONTCOVER

// x/lockup module sentinel errors.
var (
	ErrNotLockOwner                      = errorsmod.Register(ModuleName, 1, "msg sender is not the owner of specified lock")
	ErrSyntheticLockupAlreadyExists      = errorsmod.Register(ModuleName, 2, "synthetic lockup already exists for same lock and suffix")
	ErrSyntheticDurationLongerThanNative = errorsmod.Register(ModuleName, 3, "synthetic lockup duration should be shorter than native lockup duration")
	ErrLockupNotFound                    = errorsmod.Register(ModuleName, 4, "lockup not found")
)
