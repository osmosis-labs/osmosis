package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/gamm module sentinel errors
var (
	ErrPoolNotFound      = sdkerrors.Register(ModuleName, 1, "pool not found")
	ErrPoolAlreadyExist  = sdkerrors.Register(ModuleName, 2, "pool already exist")
	ErrPoolLocked        = sdkerrors.Register(ModuleName, 3, "pool is locked")
	ErrTooLittleRecords  = sdkerrors.Register(ModuleName, 4, "pool should have at least 2 records")
	ErrTooManyRecords    = sdkerrors.Register(ModuleName, 5, "pool has too many records")
	ErrLimitMaxAmount    = sdkerrors.Register(ModuleName, 6, "calculated amount is larger than max amount")
	ErrLimitMinAmount    = sdkerrors.Register(ModuleName, 7, "calculated amount is lesser than min amount")
	ErrLimitMaxPrice     = sdkerrors.Register(ModuleName, 8, "spot price exceeds max spot price")
	ErrInvalidMathApprox = sdkerrors.Register(ModuleName, 9, "invalid calculated result")
)
