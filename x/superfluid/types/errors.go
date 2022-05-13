package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/superfluid module errors.
var (
	ErrMultipleCoinsLockupNotSupported = sdkerrors.Register(ModuleName, 1, "multiple coins lockup is not supported")
	ErrUnbondingLockupNotSupported     = sdkerrors.Register(ModuleName, 2, "unbonding lockup is not allowed to participate in superfluid staking")
	ErrNotEnoughLockupDuration         = sdkerrors.Register(ModuleName, 3, "lockup does not have enough lock duration")
	ErrOsmoEquivalentZeroNotAllowed    = sdkerrors.Register(ModuleName, 4, "not able to do superfluid staking for zero osmo equivalent")
	ErrNotSuperfluidUsedLockup         = sdkerrors.Register(ModuleName, 5, "lockup is not used for superfluid staking")
	ErrSameValidatorRedelegation       = sdkerrors.Register(ModuleName, 6, "redelegation to the same validator is not allowed")
	ErrAlreadyUsedSuperfluidLockup     = sdkerrors.Register(ModuleName, 7, "lockup is already being used for superfluid staking")
	ErrUnbondingSyntheticLockupExists  = sdkerrors.Register(ModuleName, 8, "unbonding synthetic lockup exists on the validator")
	ErrBondingLockupNotSupported       = sdkerrors.Register(ModuleName, 9, "bonded superfluid stake is not allowed to have underlying lock unlocked")

	ErrNonSuperfluidAsset = sdkerrors.Register(ModuleName, 10, "provided asset is not supported for superfluid staking")

	ErrPoolNotWhitelisted   = sdkerrors.Register(ModuleName, 41, "pool not whitelisted to unpool")
	ErrLockUnpoolNotAllowed = sdkerrors.Register(ModuleName, 42, "lock not eligible for unpooling")
	ErrLockLengthMismatch   = sdkerrors.Register(ModuleName, 43, "lock has more than one asset")
)
