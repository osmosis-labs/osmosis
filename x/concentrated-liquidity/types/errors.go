package types

import (
	fmt "fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type PoolDoesNotExistError struct {
	PoolId uint64
}

func (e PoolDoesNotExistError) Error() string {
	return fmt.Sprintf("pool with ID %d does not exist", e.PoolId)
}

// x/gamm module sentinel errors.
var (
	ErrInvalidLowerUpperTick = sdkerrors.Register(ModuleName, 1, "lower tick must be lesser than upper")
	ErrInvalidLowerTick      = sdkerrors.Register(ModuleName, 2, "lower tick must be in valid range")
	ErrLimitUpperTick        = sdkerrors.Register(ModuleName, 3, "upper tick must be in valid range")

	ErrNotPositiveRequireAmount = sdkerrors.Register(ModuleName, 21, "required amount should be positive")
)
