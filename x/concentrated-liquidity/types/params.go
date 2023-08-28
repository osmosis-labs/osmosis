package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

// Parameter store keys.
var (
	KeyAuthorizedTickSpacing              = []byte("AuthorizedTickSpacing")
	KeyAuthorizedSpreadFactors            = []byte("AuthorizedSpreadFactors")
	KeyDiscountRate                       = []byte("DiscountRate")
	KeyAuthorizedQuoteDenoms              = []byte("AuthorizedQuoteDenoms")
	KeyAuthorizedUptimes                  = []byte("AuthorizedUptimes")
	KeyIsPermisionlessPoolCreationEnabled = []byte("IsPermisionlessPoolCreationEnabled")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for concentrated-liquidity module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(authorizedTickSpacing []uint64, authorizedSpreadFactors []osmomath.Dec, discountRate osmomath.Dec, authorizedQuoteDenoms []string, authorizedUptimes []time.Duration, isPermissionlessPoolCreationEnabled bool) Params {
	return Params{
		AuthorizedTickSpacing:               authorizedTickSpacing,
		AuthorizedSpreadFactors:             authorizedSpreadFactors,
		AuthorizedQuoteDenoms:               authorizedQuoteDenoms,
		BalancerSharesRewardDiscount:        discountRate,
		AuthorizedUptimes:                   authorizedUptimes,
		IsPermissionlessPoolCreationEnabled: isPermissionlessPoolCreationEnabled,
	}
}

// DefaultParams returns default concentrated-liquidity module parameters.
func DefaultParams() Params {
	return Params{
		AuthorizedTickSpacing:   AuthorizedTickSpacing,
		AuthorizedSpreadFactors: AuthorizedSpreadFactors,
		AuthorizedQuoteDenoms: []string{
			"uosmo",
			"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2", // ATOM
			"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", // DAI
			"ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", // USDC
		},
		BalancerSharesRewardDiscount:        DefaultBalancerSharesDiscount,
		AuthorizedUptimes:                   DefaultAuthorizedUptimes,
		IsPermissionlessPoolCreationEnabled: false,
	}
}

// Validate params.
func (p Params) Validate() error {
	if err := validateTicks(p.AuthorizedTickSpacing); err != nil {
		return err
	}
	if err := validateSpreadFactors(p.AuthorizedSpreadFactors); err != nil {
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
	if err := validateAuthorizedUptimes(p.AuthorizedUptimes); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthorizedTickSpacing, &p.AuthorizedTickSpacing, validateTicks),
		paramtypes.NewParamSetPair(KeyAuthorizedSpreadFactors, &p.AuthorizedSpreadFactors, validateSpreadFactors),
		paramtypes.NewParamSetPair(KeyAuthorizedQuoteDenoms, &p.AuthorizedQuoteDenoms, validateAuthorizedQuoteDenoms),
		paramtypes.NewParamSetPair(KeyIsPermisionlessPoolCreationEnabled, &p.IsPermissionlessPoolCreationEnabled, validateIsPermissionLessPoolCreationEnabled),
		paramtypes.NewParamSetPair(KeyDiscountRate, &p.BalancerSharesRewardDiscount, validateBalancerSharesDiscount),
		paramtypes.NewParamSetPair(KeyAuthorizedUptimes, &p.AuthorizedUptimes, validateAuthorizedUptimes),
	}
}

// validateTicks validates that the given parameter is a slice of strings that can be converted to unsigned 64-bit integers.
// If the parameter is not of the correct type or any of the strings cannot be converted, an error is returned.
func validateTicks(i interface{}) error {
	// Convert the given parameter to a slice of uint64s.
	authorizedTickSpacing, ok := i.([]uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Both max and min ticks must be multiple of every authorized tick spacing.
	// Otherwise, might end up running into edge cases when setting full range positions
	// and not being able to reach max and min ticks.
	for _, tickSpacing := range authorizedTickSpacing {
		if tickSpacing == 0 {
			return fmt.Errorf("tick spacing cannot be zero")
		}

		tickSpacingInt64 := int64(tickSpacing)

		if MaxTick%tickSpacingInt64 != 0 {
			return fmt.Errorf("max tick (%d) is not a multiple of tick spacing (%d)", MaxTick, tickSpacing)
		}
		if MinInitializedTick%tickSpacingInt64 != 0 {
			return fmt.Errorf("in tick (%d) is not a multiple of tick spacing (%d)", MinInitializedTick, tickSpacing)
		}

		if tickSpacingInt64 > MaxTick {
			return fmt.Errorf("tick spacing (%d) cannot be greater than max tick spacing (%d)", tickSpacing, MaxTick)
		}

		if tickSpacingInt64 < MinInitializedTick {
			return fmt.Errorf("tick spacing (%d) cannot be less than min tick spacing (%d)", tickSpacing, MinInitializedTick)
		}
	}

	return nil
}

// validateSpreadFactors validates that the given parameter is a slice of strings that can be converted to osmomath.Decs.
// If the parameter is not of the correct type or any of the strings cannot be converted, an error is returned.
func validateSpreadFactors(i interface{}) error {
	// Convert the given parameter to a slice of osmomath.Decs.
	_, ok := i.([]osmomath.Dec)
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

// validateBalancerSharesDiscount validates that the given parameter is a osmomath.Dec. Returns error if the parameter is not of the correct type.
func validateBalancerSharesDiscount(i interface{}) error {
	// Convert the given parameter to osmomath.Dec.
	balancerSharesRewardDiscount, ok := i.(osmomath.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Ensure that the passed in discount rate is between 0 and 1.
	if balancerSharesRewardDiscount.IsNegative() || balancerSharesRewardDiscount.GT(osmomath.OneDec()) {
		return InvalidDiscountRateError{DiscountRate: balancerSharesRewardDiscount}
	}

	return nil
}

// validateAuthorizedUptimes validates a slice of authorized uptimes for a given pool.
//
// Parameters:
// - i: The parameter to validate.
//
// Returns:
// - An error if given type is not duration slice.
// - An error if given slice is empty.
// - An error if any of the uptimes are invalid (i.e. not part of the list of supported uptimes).
func validateAuthorizedUptimes(i interface{}) error {
	authorizedUptimes, ok := i.([]time.Duration)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(authorizedUptimes) == 0 {
		return fmt.Errorf("authorized uptimes cannot be empty")
	}

	// Check if each passed in uptime is in the list of supported uptimes
	for _, uptime := range authorizedUptimes {
		supported := false
		for _, supportedUptime := range SupportedUptimes {
			if uptime == supportedUptime {
				supported = true

				// We break here to save on iterations
				break
			}
		}

		if !supported {
			return UptimeNotSupportedError{Uptime: uptime}
		}
	}

	return nil
}
