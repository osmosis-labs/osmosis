package types

import (
	"errors"
	"fmt"
)

type NoPoolForDenomPairError struct {
	BaseDenom  string
	MatchDenom string
}

func (e NoPoolForDenomPairError) Error() string {
	return fmt.Sprintf("highest liquidity pool between base %s and match denom %s not found", e.BaseDenom, e.MatchDenom)
}

// Is implements error matching for errors.Is to match any NoPoolForDenomPairError regardless of field values
func (e NoPoolForDenomPairError) Is(target error) bool {
	_, ok := target.(NoPoolForDenomPairError)
	return ok
}

var ErrRouteDoubleContainsPool = errors.New("cannot be trading on the same pool twice")
