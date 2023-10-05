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
	return fmt.Sprintf("desired denom (%s) was not found in the pool", e.DesiredDenom)
}
