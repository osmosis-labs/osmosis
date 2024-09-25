package types

import (
	fmt "fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
)

// Parameter store keys.
var (
	KeyAuthorizedTickSpacing              = []byte("AuthorizedTickSpacing")
	KeyAuthorizedSpreadFactors            = []byte("AuthorizedSpreadFactors")
	KeyDiscountRate                       = []byte("DiscountRate")
	KeyAuthorizedQuoteDenoms              = []byte("AuthorizedQuoteDenoms")
	KeyAuthorizedUptimes                  = []byte("AuthorizedUptimes")
	KeyIsPermisionlessPoolCreationEnabled = []byte("IsPermisionlessPoolCreationEnabled")
	KeyUnrestrictedPoolCreatorWhitelist   = []byte("UnrestrictedPoolCreatorWhitelist")
	KeyHookGasLimit                       = []byte("HookGasLimit")

	_ paramtypes.ParamSet = &Params{}
)

// ParamTable for concentrated-liquidity module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(authorizedTickSpacing []uint64, authorizedSpreadFactors []osmomath.Dec, discountRate osmomath.Dec, authorizedUptimes []time.Duration, isPermissionlessPoolCreationEnabled bool, unrestrictedPoolCreatorWhitelist []string, hookGasLimit uint64) Params {
	return Params{
		AuthorizedTickSpacing:               authorizedTickSpacing,
		AuthorizedSpreadFactors:             authorizedSpreadFactors,
		BalancerSharesRewardDiscount:        discountRate,
		AuthorizedUptimes:                   authorizedUptimes,
		IsPermissionlessPoolCreationEnabled: isPermissionlessPoolCreationEnabled,
		UnrestrictedPoolCreatorWhitelist:    unrestrictedPoolCreatorWhitelist,
		HookGasLimit:                        hookGasLimit,
	}
}

// DefaultParams returns default concentrated-liquidity module parameters.
func DefaultParams() Params {
	return Params{
		AuthorizedTickSpacing:               AuthorizedTickSpacing,
		AuthorizedSpreadFactors:             AuthorizedSpreadFactors,
		BalancerSharesRewardDiscount:        DefaultBalancerSharesDiscount,
		AuthorizedUptimes:                   DefaultAuthorizedUptimes,
		IsPermissionlessPoolCreationEnabled: false,
		UnrestrictedPoolCreatorWhitelist:    DefaultUnrestrictedPoolCreatorWhitelist,
		HookGasLimit:                        DefaultContractHookGasLimit,
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
	if err := validateIsPermissionLessPoolCreationEnabled(p.IsPermissionlessPoolCreationEnabled); err != nil {
		return err
	}
	if err := validateBalancerSharesDiscount(p.BalancerSharesRewardDiscount); err != nil {
		return err
	}
	if err := validateAuthorizedUptimes(p.AuthorizedUptimes); err != nil {
		return err
	}
	if err := osmoutils.ValidateAddressList(p.UnrestrictedPoolCreatorWhitelist); err != nil {
		return err
	}
	if err := validateHookGasLimit(p.HookGasLimit); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAuthorizedTickSpacing, &p.AuthorizedTickSpacing, validateTicks),
		paramtypes.NewParamSetPair(KeyAuthorizedSpreadFactors, &p.AuthorizedSpreadFactors, validateSpreadFactors),
		paramtypes.NewParamSetPair(KeyIsPermisionlessPoolCreationEnabled, &p.IsPermissionlessPoolCreationEnabled, validateIsPermissionLessPoolCreationEnabled),
		paramtypes.NewParamSetPair(KeyDiscountRate, &p.BalancerSharesRewardDiscount, validateBalancerSharesDiscount),
		paramtypes.NewParamSetPair(KeyAuthorizedUptimes, &p.AuthorizedUptimes, validateAuthorizedUptimes),
		paramtypes.NewParamSetPair(KeyUnrestrictedPoolCreatorWhitelist, &p.UnrestrictedPoolCreatorWhitelist, osmoutils.ValidateAddressList),
		paramtypes.NewParamSetPair(KeyHookGasLimit, &p.HookGasLimit, validateHookGasLimit),
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

	if balancerSharesRewardDiscount.IsNil() {
		return fmt.Errorf("balancer shares reward discount cannot be nil")
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

// validateHookGasLimit validates that the hook gas limit is of type uint64.
func validateHookGasLimit(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type for hook gas limit: %T", i)
	}

	return nil
}
