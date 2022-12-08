package types

import (
	fmt "fmt"
	"strconv"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyAuthorizedTickSpacing = []byte("AuthorizedTickSpacing")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(authorizedTickSpacing []string) Params {
	return Params{
		AuthorizedTickSpacing: authorizedTickSpacing,
	}
}

// DefaultParams returns default lockup module parameters.
func DefaultParams() Params {
	return Params{
		AuthorizedTickSpacing: []string{"1", "10", "60", "200"},
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validateTicks(p.AuthorizedTickSpacing); err != nil {
		return err
	}
	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthorizedTickSpacing, &p.AuthorizedTickSpacing, validateTicks),
	}
}

func validateTicks(i interface{}) error {
	ticks, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, tick := range ticks {
		_, err := strconv.ParseUint(tick, 10, 64)
		if err != nil {
			return err
		}
	}

	return nil
}
