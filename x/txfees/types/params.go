package types

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyWhitelistedFeeTokenSetters   = []byte("WhitelistedFeeTokenSetters")
	KeyFeeSwapIntermediaryDenomList = []byte("FeeSwapIntermediaryDenomList")
)

// ParamTable for txfees module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(whitelistedFeeTokenSetters []string, feeSwapIntermediaryDenomList []string) Params {
	return Params{
		WhitelistedFeeTokenSetters:   whitelistedFeeTokenSetters,
		FeeSwapIntermediaryDenomList: feeSwapIntermediaryDenomList,
	}
}

// DefaultParams are the default txfees module parameters.
func DefaultParams() Params {
	return Params{
		WhitelistedFeeTokenSetters:   []string{},
		FeeSwapIntermediaryDenomList: []string{},
	}
}

// validate params.
func (p Params) Validate() error {
	if err := osmoutils.ValidateAddressList(p.WhitelistedFeeTokenSetters); err != nil {
		return err
	}

	if err := validateFeeSwapIntermediaryDenomList(p.FeeSwapIntermediaryDenomList); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyWhitelistedFeeTokenSetters, &p.WhitelistedFeeTokenSetters, osmoutils.ValidateAddressList),
		paramtypes.NewParamSetPair(KeyFeeSwapIntermediaryDenomList, &p.FeeSwapIntermediaryDenomList, validateFeeSwapIntermediaryDenomList),
	}
}

// validateFeeSwapIntermediaryDenomList validates the fee swap intermediary denom list parameter.
func validateFeeSwapIntermediaryDenomList(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Allow empty list
	if len(v) == 0 {
		return nil
	}

	// Validate each denom (basic validation for non-empty strings)
	for _, denom := range v {
		err := sdk.ValidateDenom(denom)
		if err != nil {
			return err
		}
	}

	return nil
}
