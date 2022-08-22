package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidAmount             = sdkerrors.Register(ModuleName, 2, "invalid amount")
	ErrModuleAccountAlreadyExist = sdkerrors.Register(ModuleName, 3, "module account already exists")
	ErrModuleDoesnotExist        = sdkerrors.Register(ModuleName, 4, "module account does not exist")
	ErrInvlaidModuleAccountGiven = sdkerrors.Register(ModuleName, 5, "invalid module account given")
)
