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
	return fmt.Sprintf("insufficient liquidity requested to withdraw. Actual: (%s). Available (%s)", e.Actual, e.Available)
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
	return fmt.Sprintf("insufficient amount of token %d created. Actual: (%s). Minimum (%s)", tokenNum, e.Actual, e.Minimum)
}

type NegativeLiquidityError struct {
	Liquidity sdk.Dec
}

func (e NegativeLiquidityError) Error() string {
	return fmt.Sprintf("liquidity cannot be negative, got (%d)", e.Liquidity)
}

type DenomDuplicatedError struct {
	TokenInDenom  string
	TokenOutDenom string
}

func (e DenomDuplicatedError) Error() string {
	return fmt.Sprintf("cannot trade same denomination in (%s) and out (%s)", e.TokenInDenom, e.TokenOutDenom)
}

type AmountLessThanMinError struct {
	TokenAmount sdk.Int
	TokenMin    sdk.Int
}

func (e AmountLessThanMinError) Error() string {
	return fmt.Sprintf("token amount calculated (%s) is lesser than min amount (%s)", e.TokenAmount, e.TokenMin)
}

type AmountGreaterThanMaxError struct {
	TokenAmount sdk.Int
	TokenMax    sdk.Int
}

func (e AmountGreaterThanMaxError) Error() string {
	return fmt.Sprintf("token amount calculated (%s) is greater than max amount (%s)", e.TokenAmount, e.TokenMax)
}

type TokenInDenomNotInPoolError struct {
	TokenInDenom string
}

func (e TokenInDenomNotInPoolError) Error() string {
	return fmt.Sprintf("tokenIn (%s) does not match any asset in pool", e.TokenInDenom)
}

type TokenOutDenomNotInPoolError struct {
	TokenOutDenom string
}

func (e TokenOutDenomNotInPoolError) Error() string {
	return fmt.Sprintf("tokenOut (%s) does not match any asset in pool", e.TokenOutDenom)
}

type InvalidPriceLimitError struct {
	SqrtPriceLimit sdk.Dec
	LowerBound     sdk.Dec
	UpperBound     sdk.Dec
}

func (e InvalidPriceLimitError) Error() string {
	return fmt.Sprintf("invalid sqrt price limit given (%s), should be greater than (%s) and less than (%s)", e.SqrtPriceLimit, e.LowerBound, e.UpperBound)
}

type ZeroLiquidityError struct {
}

func (e ZeroLiquidityError) Error() string {
	return fmt.Sprintf("liquidityDelta calculated equals zero")
}
