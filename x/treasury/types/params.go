package types

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"

	"gopkg.in/yaml.v2"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyUpdateTreasuryEpochIdentifier = []byte("UpdateTreasuryEpochIdentifier")
	KeyReserveAllowableOffset        = []byte("ReserveAllowableOffset")
	KeyMaxFeeMultiplier              = []byte("MaxFeeMultiplier")
	KeyWindowShort                   = []byte("WindowShort")
	KeyWindowLong                    = []byte("WindowLong")
	KeyWindowProbation               = []byte("WindowProbation")
)

// Default parameter values
var (
	DefaultWindowShort                   = uint64(4)                     // a month
	DefaultWindowLong                    = uint64(52)                    // a year
	DefaultWindowProbation               = uint64(12)                    // 3 month
	DefaultTaxRate                       = osmomath.NewDecWithPrec(1, 2) // 0.01%
	DefaultMaxFeeMultiplier              = osmomath.NewDecWithPrec(1, 0) // 1%
	DefaultReserveAllowableOffset        = osmomath.NewDecWithPrec(5, 0) // 5%
	DefaultUpdateTreasuryEpochIdentifier = "minute"
)

var _ paramstypes.ParamSet = &Params{}

// DefaultParams creates default treasury module parameters
func DefaultParams() Params {
	return Params{
		UpdateTreasuryEpochIdentifier: DefaultUpdateTreasuryEpochIdentifier,
		ReserveAllowableOffset:        DefaultReserveAllowableOffset,
		MaxFeeMultiplier:              DefaultMaxFeeMultiplier,
		WindowShort:                   DefaultWindowShort,
		WindowLong:                    DefaultWindowLong,
		WindowProbation:               DefaultWindowProbation,
	}
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// String implements fmt.Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of treasury module's parameters.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyUpdateTreasuryEpochIdentifier, &p.UpdateTreasuryEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramstypes.NewParamSetPair(KeyWindowShort, &p.WindowShort, validateWindowShort),
		paramstypes.NewParamSetPair(KeyWindowLong, &p.WindowLong, validateWindowLong),
		paramstypes.NewParamSetPair(KeyWindowProbation, &p.WindowProbation, validateWindowProbation),
		paramstypes.NewParamSetPair(KeyReserveAllowableOffset, &p.ReserveAllowableOffset, validateReserveAllowableOffset),
		paramstypes.NewParamSetPair(KeyMaxFeeMultiplier, &p.MaxFeeMultiplier, validateMaxFeeMultiplier),
	}
}

// Validate performs basic validation on treasury parameters.
func (p Params) Validate() error {
	if epochtypes.ValidateEpochIdentifierString(p.UpdateTreasuryEpochIdentifier) != nil {
		return fmt.Errorf("treasury parameter UpdateTreasuryEpochIdentifier must be a valid epoch identifier: %s", p.UpdateTreasuryEpochIdentifier)
	}
	if p.MaxFeeMultiplier.GT(osmomath.NewDecWithPrec(10, 0)) {
		return fmt.Errorf("treasury parameter MaxFeeMultiplier must be lower than 10: %s", p.MaxFeeMultiplier)
	}
	if p.WindowLong <= p.WindowShort {
		return fmt.Errorf("treasury parameter WindowLong must be bigger than WindowShort: (%d, %d)", p.WindowLong, p.WindowShort)
	}

	return nil
}

func validateReserveAllowableOffset(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("reserve allowable offset must be positive: %s", v)
	}

	return nil
}

func validateMaxFeeMultiplier(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max fee multiplier must be positive: %s", v)
	}

	return nil
}

func validateWindowShort(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateWindowLong(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateWindowProbation(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
