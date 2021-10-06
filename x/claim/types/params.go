package types

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyClaimDenom         = []byte("ClaimDenom")
	KeyDurationUntilDecay = []byte("DurationUntilDecay")
	KeyDurationOfDecay    = []byte("DurationOfDecay")
)

// ParamTable for claim module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

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

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyClaimDenom, &p.ClaimDenom, validateClaimDenom),
		paramtypes.NewParamSetPair(KeyClaimDenom, &p.DurationUntilDecay, validateDuration),
		paramtypes.NewParamSetPair(KeyClaimDenom, &p.DurationUntilDecay, validateClaimDenom),
	}
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
