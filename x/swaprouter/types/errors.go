package types

import (
	"errors"
	"fmt"
)

type nonPositiveAmountErr struct {
	Amount string
}

func (e nonPositiveAmountErr) Error() string {
	return fmt.Sprintf("min out amount or max in amount should be positive, was (%s)", e.Amount)
}

var (
	ErrEmptyRoutes = errors.New("provided empty routes")
)
