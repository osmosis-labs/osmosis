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
