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

var ErrRouteDoubleContainsPool = errors.New("cannot be trading on the same pool twice")
