package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoFarmExist       = sdkerrors.Register(ModuleName, 1, "farm doesn't exist")
	ErrNoFarmerExist     = sdkerrors.Register(ModuleName, 2, "farmer doesn't exist")
	ErrInsufficientShare = sdkerrors.Register(ModuleName, 3, "insufficient share")
)
