package types

import (
	"errors"
	fmt "fmt"
)

// x/concentrated-liquidity module sentinel errors.
type ErrInvalidLowerUpperTick struct {
	LowerTick int64
	UpperTick int64
}

func (e ErrInvalidLowerUpperTick) Error() string {
	return fmt.Sprintf("Lower tick must be lesser than upper. Got lower: %d, upper: %d", e.LowerTick, e.UpperTick)
}

var (
	ErrInvalidLowerTick      	= errors.New("lower tick must be in valid range")
	ErrLimitUpperTick        	= errors.New("upper tick must be in valid range")
	ErrNotPositiveRequireAmount = errors.New("required amount should be positive")
)
