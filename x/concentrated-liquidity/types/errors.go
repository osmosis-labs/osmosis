package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// x/concentrated-liquidity module sentinel errors.
type InvalidLowerUpperTickError struct {
	LowerTick int64
	UpperTick int64
}

func (e InvalidLowerUpperTickError) Error() string {
	return fmt.Sprintf("Lower tick must be lesser than upper. Got lower: %d, upper: %d", e.LowerTick, e.UpperTick)
}

type InvalidLowerTickError struct {
	LowerTick int64
}

func (e InvalidLowerTickError) Error() string {
	return fmt.Sprintf("Lower tick must be in range [%d, %d]. Got: %d", MinTick, MaxTick, e.LowerTick)
}

type InvalidUpperTickError struct {
	UpperTick int64
}

func (e InvalidUpperTickError) Error() string {
	return fmt.Sprintf("Upper tick must be in range [%d, %d]. Got: %d", MinTick, MaxTick, e.UpperTick)
}

type NotPositiveRequireAmountError struct {
	Amount string
}

func (e NotPositiveRequireAmountError) Error() string {
	return fmt.Sprintf("Required amount should be positive. Got: %s", e.Amount)
}

type PositionNotFoundError struct {
	PoolId    uint64
	LowerTick int64
	UpperTick int64
}

func (e PositionNotFoundError) Error() string {
	return fmt.Sprintf("position not found. pool id (%d), lower tick (%d), upper tick (%d)", e.PoolId, e.LowerTick, e.UpperTick)
}

type PoolNotFoundError struct {
	PoolId uint64
}

func (e PoolNotFoundError) Error() string {
	return fmt.Sprintf("pool not found. pool id (%d)", e.PoolId)
}

type InvalidTickError struct {
	Tick    int64
	IsLower bool
}

func (e InvalidTickError) Error() string {
	tickStr := "upper"
	if e.IsLower {
		tickStr = "lower"
	}
	return fmt.Sprintf("%s tick (%d) is invalid, Must be >= %d and <= %d", tickStr, e.Tick, MinTick, MaxTick)
}

type InsufficientLiquidityError struct {
	Actual    sdk.Dec
	Available sdk.Dec
}

func (e InsufficientLiquidityError) Error() string {
	return fmt.Sprintf("insufficient liqudity requested to withdraw. Actual: (%s). Available (%s)", e.Actual, e.Available)
}

type InsufficientLiquidityCreatedError struct {
	Actual      sdk.Int
	Minimum     sdk.Int
	IsTokenZero bool
}

func (e InsufficientLiquidityCreatedError) Error() string {
	tokenNum := uint8(0)
	if !e.IsTokenZero {
		tokenNum = 1
	}
	return fmt.Sprintf("insufficient amount of token%d created. Actual: (%s). Minimum (%s)", tokenNum, e.Actual, e.Minimum)
}

type PoolDoesNotExistError struct {
	PoolId uint64
}

func (e PoolDoesNotExistError) Error() string {
	return fmt.Sprintf("cannot initialize or update a tick for a non-existing pool. pool id (%d)", e.PoolId)
}

type InvalidPriceLimitError struct {
	SqrtPriceLimit sdk.Dec
	LowerBound     sdk.Dec
	UpperBound     sdk.Dec
}

func (e InvalidPriceLimitError) Error() string {
	return fmt.Sprintf("invalid sqrt price limit given (%s), should be greater than (%s) and less than (%s)", e.SqrtPriceLimit, e.LowerBound, e.UpperBound)
}
