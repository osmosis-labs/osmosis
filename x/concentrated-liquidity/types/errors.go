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
	ErrLimitMaxTick          = sdkerrors.Register(ModuleName, 2, "upper tick is larger than max tick")
	ErrLimitMinTick          = sdkerrors.Register(ModuleName, 3, "lower tick is lesser than min tick")

	ErrNotPositiveRequireAmount = sdkerrors.Register(ModuleName, 21, "required amount should be positive")
)
