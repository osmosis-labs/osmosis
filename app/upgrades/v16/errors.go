package v16

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrMustHaveTwoDenoms = errors.New("can only have 2 denoms in CL pool")
)

type NoDesiredDenomInPoolError struct {
	DesiredDenom string
}

func (e NoDesiredDenomInPoolError) Error() string {
	return fmt.Sprintf("no desired denom in pool (%s)", e.DesiredDenom)
}

type CouldNotFindGaugeToRedirectError struct {
	DistributionEpochDuration time.Duration
}

func (e CouldNotFindGaugeToRedirectError) Error() string {
	return fmt.Sprintf("could not find gauge for distribution epoch duration %d", e.DistributionEpochDuration)
}
