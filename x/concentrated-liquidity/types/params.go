package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyAuthorizedTickSpacing = []byte("AuthorizedTickSpacing")
	KeyAuthorizedSwapFees    = []byte("AuthorizedSwapFees")
	KeyDiscountRate          = []byte("DiscountRate")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for concentrated-liquidity module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(authorizedTickSpacing []uint64, authorizedSwapFees []sdk.Dec, discountRate sdk.Dec) Params {
	return Params{
		AuthorizedTickSpacing: authorizedTickSpacing,
		AuthorizedSwapFees:    authorizedSwapFees,
		DiscountRate:          discountRate,
	}
}

// DefaultParams returns default concentrated-liquidity module parameters.
// TODO: Decide on what these should be initially.
// https://github.com/osmosis-labs/osmosis/issues/3684
func DefaultParams() Params {
	return Params{
		AuthorizedTickSpacing: AuthorizedTickSpacing,
		AuthorizedSwapFees: []sdk.Dec{sdk.ZeroDec(),
			sdk.MustNewDecFromStr("0.0001"),
			sdk.MustNewDecFromStr("0.0003"),
			sdk.MustNewDecFromStr("0.0005"),
			sdk.MustNewDecFromStr("0.003"),
			sdk.MustNewDecFromStr("0.01")},
		DiscountRate:         DefaultDiscountRate,
	}
}

// Validate params.
func (p Params) Validate() error {
	if err := validateTicks(p.AuthorizedTickSpacing); err != nil {
		return err
	}
	if err := validateSwapFees(p.AuthorizedSwapFees); err != nil {
		return err
	}
	if err := validateDiscountRate(p.DiscountRate); err!= nil {
        return err
    }
	return nil
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthorizedTickSpacing, &p.AuthorizedTickSpacing, validateTicks),
		paramtypes.NewParamSetPair(KeyAuthorizedSwapFees, &p.AuthorizedSwapFees, validateSwapFees),
		paramtypes.NewParamSetPair(KeyDiscountRate, &p.DiscountRate, validateDiscountRate),
	}
}

// validateTicks validates that the given parameter is a slice of strings that can be converted to unsigned 64-bit integers.
// If the parameter is not of the correct type or any of the strings cannot be converted, an error is returned.
func validateTicks(i interface{}) error {
	// Convert the given parameter to a slice of uint64s.
	_, ok := i.([]uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// validateSwapFees validates that the given parameter is a slice of strings that can be converted to sdk.Decs.
// If the parameter is not of the correct type or any of the strings cannot be converted, an error is returned.
func validateSwapFees(i interface{}) error {
	// Convert the given parameter to a slice of sdk.Decs.
	_, ok := i.([]sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// validateDiscountRate validates that the given parameter is a sdk.Dec. Returns error if the parameter is not of the correc type.
func validateDiscountRate(i interface{}) error {
	// Convert the given parameter to sdk.Dec.
	_, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
