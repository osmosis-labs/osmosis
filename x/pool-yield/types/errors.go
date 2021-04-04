package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoFarmIdExist = sdkerrors.Register(ModuleName, 1, "no farm id exist")
)
