package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(distrEpochIdentifier string) Params {
	return Params{
		DistrEpochIdentifier: distrEpochIdentifier,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier: "week",
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateDistrEpochIdentifier(p.DistrEpochIdentifier); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, validateDistrEpochIdentifier),
	}
}

func validateDistrEpochIdentifier(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == "" {
		return fmt.Errorf("empty distribution epoch identifier: %+v", i)
	}

	return nil
}
