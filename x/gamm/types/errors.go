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
	ErrPoolNotFound              = sdkerrors.Register(ModuleName, 1, "pool not found")
	ErrPoolAlreadyExist          = sdkerrors.Register(ModuleName, 2, "pool already exist")
	ErrPoolLocked                = sdkerrors.Register(ModuleName, 3, "pool is locked")
	ErrTooFewPoolAssets          = sdkerrors.Register(ModuleName, 4, "pool should have at least 2 assets, as they must be swapping between at least two assets")
	ErrTooManyPoolAssets         = sdkerrors.Register(ModuleName, 5, "pool has too many assets (currently capped at 8 assets per balancer pool and 2 per stableswap)")
	ErrLimitMaxAmount            = sdkerrors.Register(ModuleName, 6, "calculated amount is larger than max amount")
	ErrLimitMinAmount            = sdkerrors.Register(ModuleName, 7, "calculated amount is lesser than min amount")
	ErrInvalidMathApprox         = sdkerrors.Register(ModuleName, 8, "invalid calculated result")
	ErrAlreadyInvalidPool        = sdkerrors.Register(ModuleName, 9, "destruction on already invalid pool")
	ErrInvalidPool               = sdkerrors.Register(ModuleName, 10, "attempting to create an invalid pool")
	ErrDenomNotFoundInPool       = sdkerrors.Register(ModuleName, 11, "denom does not exist in pool")
	ErrDenomAlreadyInPool        = sdkerrors.Register(ModuleName, 12, "denom already exists in the pool")

	ErrEmptyRoutes              = sdkerrors.Register(ModuleName, 21, "routes not defined")
	ErrEmptyPoolAssets          = sdkerrors.Register(ModuleName, 22, "PoolAssets not defined")
	ErrNegativeSwapFee          = sdkerrors.Register(ModuleName, 23, "swap fee is negative")
	ErrNegativeExitFee          = sdkerrors.Register(ModuleName, 24, "exit fee is negative")
	ErrTooMuchSwapFee           = sdkerrors.Register(ModuleName, 25, "swap fee should be lesser than 1 (100%)")
	ErrTooMuchExitFee           = sdkerrors.Register(ModuleName, 26, "exit fee should be lesser than 1 (100%)")
	ErrNotPositiveWeight        = sdkerrors.Register(ModuleName, 27, "token weight should be greater than 0")
	ErrWeightTooLarge           = sdkerrors.Register(ModuleName, 28, "user specified token weight should be less than 2^20")
	ErrNotPositiveCriteria      = sdkerrors.Register(ModuleName, 29, "min out amount or max in amount should be positive")
	ErrNotPositiveRequireAmount = sdkerrors.Register(ModuleName, 30, "required amount should be positive")
	ErrTooManyTokensOut         = sdkerrors.Register(ModuleName, 31, "tx is trying to get more tokens out of the pool than exist")
	ErrSpotPriceOverflow        = sdkerrors.Register(ModuleName, 32, "invalid spot price (overflowed)")
	ErrSpotPriceInternal        = sdkerrors.Register(ModuleName, 33, "internal spot price error")

	ErrPoolParamsInvalidDenom     = sdkerrors.Register(ModuleName, 50, "pool params' LBP params has an invalid denomination")
	ErrPoolParamsInvalidNumDenoms = sdkerrors.Register(ModuleName, 51, "pool params' LBP doesn't have same number of params as underlying pool")

	ErrNotImplemented = sdkerrors.Register(ModuleName, 60, "function not implemented")

	ErrNotStableSwapPool               = sdkerrors.Register(ModuleName, 61, "not stableswap pool")
	ErrInvalidStableswapScalingFactors = sdkerrors.Register(ModuleName, 62, "length between liquidity and scaling factors mismatch")
	ErrNotScalingFactorGovernor        = sdkerrors.Register(ModuleName, 63, "not scaling factor governor")
	ErrInvalidScalingFactors           = sdkerrors.Register(ModuleName, 64, "invalid scaling factor")
)
