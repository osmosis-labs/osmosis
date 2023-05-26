package types

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PoolDoesNotExistError struct {
	PoolId uint64
}

func (e PoolDoesNotExistError) Error() string {
	return fmt.Sprintf("pool with ID %d does not exist", e.PoolId)
}

type UnsortedPoolLiqError struct {
	ActualLiquidity sdk.Coins
}

func (e UnsortedPoolLiqError) Error() string {
	return fmt.Sprintf(`unsorted initial pool liquidity: %s. 
	Please sort and make sure scaling factor order matches initial liquidity coin order`, e.ActualLiquidity)
}

type LiquidityAndScalingFactorCountMismatchError struct {
	LiquidityCount     int
	ScalingFactorCount int
}

func (e LiquidityAndScalingFactorCountMismatchError) Error() string {
	return fmt.Sprintf("liquidity count (%d) must match scaling factor count (%d)", e.LiquidityCount, e.ScalingFactorCount)
}

type ConcentratedPoolMigrationLinkNotFoundError struct {
	PoolIdLeaving uint64
}

func (e ConcentratedPoolMigrationLinkNotFoundError) Error() string {
	return fmt.Sprintf("given poolIdLeaving (%d) does not have a canonical link for any concentrated pool", e.PoolIdLeaving)
}

type BalancerPoolMigrationLinkNotFoundError struct {
	PoolIdEntering uint64
}

func (e BalancerPoolMigrationLinkNotFoundError) Error() string {
	return fmt.Sprintf("given PoolIdEntering (%d) does not have a canonical link for any balancer pool", e.PoolIdEntering)
}

// x/gamm module sentinel errors.
var (
	ErrPoolNotFound        = errorsmod.Register(ModuleName, 1, "pool not found")
	ErrPoolAlreadyExist    = errorsmod.Register(ModuleName, 2, "pool already exist")
	ErrPoolLocked          = errorsmod.Register(ModuleName, 3, "pool is locked")
	ErrTooFewPoolAssets    = errorsmod.Register(ModuleName, 4, "pool should have at least 2 assets, as they must be swapping between at least two assets")
	ErrTooManyPoolAssets   = errorsmod.Register(ModuleName, 5, "pool has too many assets (currently capped at 8 assets for both balancer and stableswap)")
	ErrLimitMaxAmount      = errorsmod.Register(ModuleName, 6, "calculated amount is larger than max amount")
	ErrLimitMinAmount      = errorsmod.Register(ModuleName, 7, "calculated amount is lesser than min amount")
	ErrInvalidMathApprox   = errorsmod.Register(ModuleName, 8, "invalid calculated result")
	ErrAlreadyInvalidPool  = errorsmod.Register(ModuleName, 9, "destruction on already invalid pool")
	ErrInvalidPool         = errorsmod.Register(ModuleName, 10, "attempting to create an invalid pool")
	ErrDenomNotFoundInPool = errorsmod.Register(ModuleName, 11, "denom does not exist in pool")
	ErrDenomAlreadyInPool  = errorsmod.Register(ModuleName, 12, "denom already exists in the pool")

	ErrEmptyRoutes              = errorsmod.Register(ModuleName, 21, "routes not defined")
	ErrEmptyPoolAssets          = errorsmod.Register(ModuleName, 22, "PoolAssets not defined")
	ErrNegativeSpreadFactor     = errorsmod.Register(ModuleName, 23, "spread factor is negative")
	ErrNegativeExitFee          = errorsmod.Register(ModuleName, 24, "exit fee is negative")
	ErrTooMuchSpreadFactor      = errorsmod.Register(ModuleName, 25, "spread factor should be lesser than 1 (100%)")
	ErrTooMuchExitFee           = errorsmod.Register(ModuleName, 26, "exit fee should be lesser than 1 (100%)")
	ErrNotPositiveWeight        = errorsmod.Register(ModuleName, 27, "token weight should be greater than 0")
	ErrWeightTooLarge           = errorsmod.Register(ModuleName, 28, "user specified token weight should be less than 2^20")
	ErrNotPositiveCriteria      = errorsmod.Register(ModuleName, 29, "min out amount or max in amount should be positive")
	ErrNotPositiveRequireAmount = errorsmod.Register(ModuleName, 30, "required amount should be positive")
	ErrTooManyTokensOut         = errorsmod.Register(ModuleName, 31, "tx is trying to get more tokens out of the pool than exist")
	ErrSpotPriceOverflow        = errorsmod.Register(ModuleName, 32, "invalid spot price (overflowed)")
	ErrSpotPriceInternal        = errorsmod.Register(ModuleName, 33, "internal spot price error")

	ErrPoolParamsInvalidDenom     = errorsmod.Register(ModuleName, 50, "pool params' LBP params has an invalid denomination")
	ErrPoolParamsInvalidNumDenoms = errorsmod.Register(ModuleName, 51, "pool params' LBP doesn't have same number of params as underlying pool")

	ErrNotImplemented = errorsmod.Register(ModuleName, 60, "function not implemented")

	ErrNotStableSwapPool          = errorsmod.Register(ModuleName, 61, "not stableswap pool")
	ErrInvalidScalingFactorLength = errorsmod.Register(ModuleName, 62, "pool liquidity and scaling factors must have same length")
	ErrNotScalingFactorGovernor   = errorsmod.Register(ModuleName, 63, "not scaling factor governor")
	ErrInvalidScalingFactors      = errorsmod.Register(ModuleName, 64, "scaling factors cannot be 0 or use more than 63 bits")
	ErrHitMaxScaledAssets         = errorsmod.Register(ModuleName, 65, "post-scaled pool assets can not exceed 10^34")
	ErrHitMinScaledAssets         = errorsmod.Register(ModuleName, 66, "post-scaled pool assets can not be less than 1")
)
