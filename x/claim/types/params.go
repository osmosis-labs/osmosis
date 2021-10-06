package types

import (
	"fmt"
	"time"
)

func NewParams(claimDenom string, durationUntilDecay time.Duration, durationOfDecay time.Duration) Params {
	return Params{
		ClaimDenom:         claimDenom,
		DurationUntilDecay: durationUntilDecay,
		DurationOfDecay:    durationOfDecay,
	}
}

// default claim module parameters
func DefaultParams() Params {
	return Params{
		ClaimDenom:         "uosmo",
		DurationUntilDecay: time.Hour,
		DurationOfDecay:    time.Hour * 5,
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateClaimDenom(p.ClaimDenom); err != nil {
		return err
	}

	if err := validateDuration(p.DurationOfDecay); err != nil {
		return err
	}

	if err := validateDuration(p.DurationUntilDecay); err != nil {
		return err
	}

	return nil
}

func validateClaimDenom(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDuration(i interface{}) error {
	_, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
