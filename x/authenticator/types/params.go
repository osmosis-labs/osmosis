package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		MaximumUnauthenticatedGas: 20000,
		AuthenticatorActiveState:  true,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaximumUnauthenticatedGas, &p.MaximumUnauthenticatedGas, validateMaximumUnauthenticatedGas),
		paramtypes.NewParamSetPair(KeyAuthenticatorActiveState, &p.AuthenticatorActiveState, validateAuthenticatorActiveState),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// Validate Default Gas Reduction
func validateMaximumUnauthenticatedGas(i interface{}) error {
	// Convert the given parameter to a uint64.
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateAuthenticatorActiveState(i interface{}) error {
	// Convert the given parameter to a bool.
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
