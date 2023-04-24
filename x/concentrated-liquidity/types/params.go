package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyAuthorizedTickSpacing              = []byte("AuthorizedTickSpacing")
	KeyAuthorizedSwapFees                 = []byte("AuthorizedSwapFees")
	KeyDiscountRate                       = []byte("DiscountRate")
	KeyAuthorizedQuoteDenoms              = []byte("AuthorizedQuoteDenoms")
	KeyIsPermisionlessPoolCreationEnabled = []byte("IsPermisionlessPoolCreationEnabled")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for concentrated-liquidity module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(authorizedTickSpacing []uint64, authorizedSwapFees []sdk.Dec, discountRate sdk.Dec, authorizedQuoteDenoms []string, isPermissionlessPoolCreationEnabled bool) Params {
	return Params{
		AuthorizedTickSpacing:               authorizedTickSpacing,
		AuthorizedSwapFees:                  authorizedSwapFees,
		AuthorizedQuoteDenoms:               authorizedQuoteDenoms,
		BalancerSharesRewardDiscount:        discountRate,
		IsPermissionlessPoolCreationEnabled: isPermissionlessPoolCreationEnabled,
	}
}

// DefaultParams returns default concentrated-liquidity module parameters.
// TODO: Decide on what these should be initially.
// https://github.com/osmosis-labs/osmosis/issues/3684
func DefaultParams() Params {
	return Params{
		AuthorizedTickSpacing: AuthorizedTickSpacing,
		AuthorizedSwapFees: []sdk.Dec{
			sdk.ZeroDec(),
			sdk.MustNewDecFromStr("0.0001"),
			sdk.MustNewDecFromStr("0.0003"),
			sdk.MustNewDecFromStr("0.0005"),
			sdk.MustNewDecFromStr("0.002"),
			sdk.MustNewDecFromStr("0.01")},
		AuthorizedQuoteDenoms: []string{
			"uosmo",
			"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", // DAI
			"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
		},
		IsPermissionlessPoolCreationEnabled: false,
		BalancerSharesRewardDiscount:        DefaultBalancerSharesDiscount,
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
	if err := validateAuthorizedQuoteDenoms(p.AuthorizedQuoteDenoms); err != nil {
		return err
	}
	if err := validateIsPermissionLessPoolCreationEnabled(p.IsPermissionlessPoolCreationEnabled); err != nil {
		return err
	}
	if err := validateBalancerSharesDiscount(p.BalancerSharesRewardDiscount); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthorizedTickSpacing, &p.AuthorizedTickSpacing, validateTicks),
		paramtypes.NewParamSetPair(KeyAuthorizedSwapFees, &p.AuthorizedSwapFees, validateSwapFees),
		paramtypes.NewParamSetPair(KeyAuthorizedQuoteDenoms, &p.AuthorizedQuoteDenoms, validateAuthorizedQuoteDenoms),
		paramtypes.NewParamSetPair(KeyIsPermisionlessPoolCreationEnabled, &p.IsPermissionlessPoolCreationEnabled, validateIsPermissionLessPoolCreationEnabled),
		paramtypes.NewParamSetPair(KeyDiscountRate, &p.BalancerSharesRewardDiscount, validateBalancerSharesDiscount),
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

// validateAuthorizedQuoteDenoms validates a slice of authorized quote denoms.
//
// Parameters:
// - i: The parameter to validate.
//
// Returns:
// - An error if given type is not string slice.
// - An error if given slice is empty.
// - An error if any of the denoms are invalid.
func validateAuthorizedQuoteDenoms(i interface{}) error {
	authorizedQuoteDenoms, ok := i.([]string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(authorizedQuoteDenoms) == 0 {
		return fmt.Errorf("authorized quote denoms cannot be empty")
	}

	for _, denom := range authorizedQuoteDenoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return err
		}
	}

	return nil
}

// validateIsPermissionLessPoolCreationEnabled validates that the given parameter is a bool.
func validateIsPermissionLessPoolCreationEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type for is permissionless pool creation enabled flag: %T", i)
	}

	return nil
}

// validateBalancerSharesDiscount validates that the given parameter is a sdk.Dec. Returns error if the parameter is not of the correct type.
func validateBalancerSharesDiscount(i interface{}) error {
	// Convert the given parameter to sdk.Dec.
	balancerSharesRewardDiscount, ok := i.(sdk.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Ensure that the passed in discount rate is between 0 and 1.
	if balancerSharesRewardDiscount.LT(sdk.ZeroDec()) && balancerSharesRewardDiscount.GT(sdk.OneDec()) {
		return InvalidDiscountRateError{DiscountRate: balancerSharesRewardDiscount}
	}

	return nil
}
