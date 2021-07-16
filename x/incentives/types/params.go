package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
	KeyMinAutostakingRate   = []byte("MinAutostakingRate")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(distrEpochIdentifier string, minAutostakingRate sdk.Dec) Params {
	return Params{
		DistrEpochIdentifier: distrEpochIdentifier,
		MinAutostakingRate:   minAutostakingRate,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier: "week",
		MinAutostakingRate:   sdk.NewDecWithPrec(5, 1), // 50%
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateDistrEpochIdentifier(p.DistrEpochIdentifier); err != nil {
		return err
	}
	if err := validateMinAutostakingRate(p.MinAutostakingRate); err != nil {
		return err
	}

	return nil

}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, validateDistrEpochIdentifier),
		paramtypes.NewParamSetPair(KeyMinAutostakingRate, &p.MinAutostakingRate, validateMinAutostakingRate),
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

func validateMinAutostakingRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("empty auto-staking: %+v", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("negative auto-staking rate: %+v", i)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("more than 1 min auto-staking rate: %+v", i)
	}

	return nil
}
