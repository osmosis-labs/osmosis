package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrAmountNilOrZero           = sdkerrors.Register(ModuleName, 2, "amount cannot be nil or zero")
	ErrModuleAccountAlreadyExist = sdkerrors.Register(ModuleName, 3, "module account already exists")
	ErrModuleDoesnotExist        = sdkerrors.Register(ModuleName, 4, "module account does not exist")
)
