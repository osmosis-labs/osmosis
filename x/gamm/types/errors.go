package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// x/gamm module sentinel errors
var (
	ErrPoolNotFound      = sdkerrors.Register(ModuleName, 1, "pool not found")
	ErrPoolAlreadyExist  = sdkerrors.Register(ModuleName, 2, "pool already exist")
	ErrPoolLocked        = sdkerrors.Register(ModuleName, 3, "pool is locked")
	ErrTooFewPoolAssets  = sdkerrors.Register(ModuleName, 4, "pool should have at least 2 PoolAssets, as they must be swapping between at least two assets")
	ErrTooManyPoolAssets = sdkerrors.Register(ModuleName, 5, "pool has too many PoolAssets")
	ErrLimitMaxAmount    = sdkerrors.Register(ModuleName, 6, "calculated amount is larger than max amount")
	ErrLimitMinAmount    = sdkerrors.Register(ModuleName, 7, "calculated amount is lesser than min amount")
	ErrInvalidMathApprox = sdkerrors.Register(ModuleName, 8, "invalid calculated result")

	ErrEmptyRoutes              = sdkerrors.Register(ModuleName, 21, "routes not defined")
	ErrEmptyPoolAssets          = sdkerrors.Register(ModuleName, 22, "PoolAssets not defined")
	ErrNegativeSwapFee          = sdkerrors.Register(ModuleName, 23, "swap fee is negative")
	ErrNegativeExitFee          = sdkerrors.Register(ModuleName, 24, "exit fee is negative")
	ErrTooMuchSwapFee           = sdkerrors.Register(ModuleName, 25, "swap fee should be lesser than 1 (100%)")
	ErrTooMuchExitFee           = sdkerrors.Register(ModuleName, 26, "exit fee should be lesser than 1 (100%)")
	ErrNotPositiveWeight        = sdkerrors.Register(ModuleName, 27, "token weight should be positive")
	ErrNotPositiveCriteria      = sdkerrors.Register(ModuleName, 28, "min out amount or max in amount should be positive")
	ErrNotPositiveRequireAmount = sdkerrors.Register(ModuleName, 29, "required amount should be positive")
)
