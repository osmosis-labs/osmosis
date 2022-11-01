package types

import (
	fmt "fmt"
)

// x/concentrated-liquidity module sentinel errors.
var (
	ErrInvalidLowerUpperTick = fmt.Errorf("lower tick must be lesser than upper")
	ErrInvalidLowerTick      = fmt.Errorf("lower tick must be in valid range")
	ErrLimitUpperTick        = fmt.Errorf("upper tick must be in valid range")

	ErrNotPositiveRequireAmount = fmt.Errorf("required amount should be positive")
)
