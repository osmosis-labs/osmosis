package v16

import (
	"errors"
	"fmt"
)

var (
	ErrMustHaveTwoDenoms = errors.New("can only have 2 denoms in CL pool")
	ErrNoGaugeToRedirect = errors.New("could not find gauge to redirect")
)

type NoDesiredDenomInPoolError struct {
	DesiredDenom string
}

func (e NoDesiredDenomInPoolError) Error() string {
	return fmt.Sprintf("no desired denom in pool (%s)", e.DesiredDenom)
}
