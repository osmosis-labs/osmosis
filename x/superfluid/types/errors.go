package types

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
)

// x/superfluid module errors.
var (
	ErrMultipleCoinsLockupNotSupported = errorsmod.Register(ModuleName, 1, "multiple coins lockup is not supported")
	ErrUnbondingLockupNotSupported     = errorsmod.Register(ModuleName, 2, "unbonding lockup is not allowed to participate in superfluid staking")
	ErrNotEnoughLockupDuration         = errorsmod.Register(ModuleName, 3, "lockup does not have enough lock duration")
	ErrOsmoEquivalentZeroNotAllowed    = errorsmod.Register(ModuleName, 4, "not able to do superfluid staking for zero osmo equivalent")
	ErrNotSuperfluidUsedLockup         = errorsmod.Register(ModuleName, 5, "lockup is not used for superfluid staking")
	ErrSameValidatorRedelegation       = errorsmod.Register(ModuleName, 6, "redelegation to the same validator is not allowed")
	ErrAlreadyUsedSuperfluidLockup     = errorsmod.Register(ModuleName, 7, "lockup is already being used for superfluid staking")
	ErrUnbondingSyntheticLockupExists  = errorsmod.Register(ModuleName, 8, "unbonding synthetic lockup exists on the validator")
	ErrBondingLockupNotSupported       = errorsmod.Register(ModuleName, 9, "bonded superfluid stake is not allowed to have underlying lock unlocked")

	ErrNonSuperfluidAsset = errorsmod.Register(ModuleName, 10, "provided asset is not supported for superfluid staking")

	ErrPoolNotWhitelisted   = errorsmod.Register(ModuleName, 41, "pool not whitelisted to unpool")
	ErrLockUnpoolNotAllowed = errorsmod.Register(ModuleName, 42, "lock not eligible for unpooling")
	ErrLockLengthMismatch   = errorsmod.Register(ModuleName, 43, "lock has more than one asset")
)

type PositionNotSuperfluidStakedError struct {
	PositionId uint64
}

func (e PositionNotSuperfluidStakedError) Error() string {
	return fmt.Sprintf("Cannot add to position ID %d as it is not superfluid staked.", e.PositionId)
}

type LockImproperStateError struct {
	LockId            uint64
	UnbondingDuration string
}

func (e LockImproperStateError) Error() string {
	return fmt.Sprintf("lock ID %d must be bonded for %s and not unbonding.", e.LockId, e.UnbondingDuration)
}

type LockOwnerMismatchError struct {
	LockId        uint64
	LockOwner     string
	ProvidedOwner string
}

func (e LockOwnerMismatchError) Error() string {
	return fmt.Sprintf("lock ID %d owner %s does not match provided owner %s.", e.LockId, e.LockOwner, e.ProvidedOwner)
}
