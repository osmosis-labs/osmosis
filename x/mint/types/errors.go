package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrAmountNilOrZero           = errorsmod.Register(ModuleName, 2, "amount cannot be nil or zero")
	ErrModuleAccountAlreadyExist = errorsmod.Register(ModuleName, 3, "module account already exists")
	ErrModuleDoesnotExist        = errorsmod.Register(ModuleName, 4, "module account does not exist")
)
