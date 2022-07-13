package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrAmountCannotBeNilOrZero               = sdkerrors.Register(ModuleName, 1, "amount cannot be nil or zero")
	ErrDevVestingModuleAccountAlreadyCreated = sdkerrors.Register(ModuleName, 2, "module account already exists")
	ErrDevVestingModuleAccountNotCreated     = sdkerrors.Register(ModuleName, 3, "module account does not exist")
)
