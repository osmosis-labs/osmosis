package types

import (
	fmt "fmt"
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
