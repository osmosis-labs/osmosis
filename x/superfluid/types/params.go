package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyMinimumRiskFactor     = []byte("MinimumRiskFactor")
	defaultMinimumRiskFactor = sdk.NewDecWithPrec(5, 1) // 50%
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(minimumRiskFactor sdk.Dec) Params {
	return Params{
		MinimumRiskFactor: minimumRiskFactor,
	}
}

// default minting module parameters.
func DefaultParams() Params {
	return Params{
		MinimumRiskFactor: defaultMinimumRiskFactor, // 5%
	}
}

// validate params.
func (p Params) Validate() error {
	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinimumRiskFactor, &p.MinimumRiskFactor, ValidateMinimumRiskFactor),
	}
}

func ValidateMinimumRiskFactor(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.ZeroDec()) || v.GT(sdk.NewDec(100)) {
		return fmt.Errorf("minimum risk factor should be between 0 - 100: %s", v.String())
	}

	return nil
}

func ValidateUnbondingDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("unbonding duration should be positive: %s", v.String())
	}

	return nil
}
