package types

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"

	"gopkg.in/yaml.v2"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v27/app/params"
)

// Parameter keys
var (
	// The period required to recover BasePool
	KeyPoolRecoveryPeriod = []byte("PoolRecoveryPeriod")
	// Min spread
	KeyMinStabilitySpread = []byte("MinStabilitySpread")
)

// Default parameter values
var (
	DefaultBasePool           = osmomath.NewDec(1000000 * params.MicroUnit) // 1000,000sdr = 1000,000,000,000usdr
	DefaultMinStabilitySpread = osmomath.NewDecWithPrec(25, 4)              // 0.25%
)

var _ paramstypes.ParamSet = &Params{}

// DefaultParams creates default market module parameters
func DefaultParams() Params {
	return Params{
		MinStabilitySpread: DefaultMinStabilitySpread,
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
// pairs of market module's parameters.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyMinStabilitySpread, &p.MinStabilitySpread, validateMinStabilitySpread),
	}
}

// Validate a set of params
func (p Params) Validate() error {
	if p.ExchangePool.IsNegative() {
		return fmt.Errorf("exchange pool should be positive or zero, is %s", p.ExchangePool)
	}

	return nil
}

func validatePoolRecoveryPeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("pool recovery period must be positive: %d", v)
	}

	return nil
}

func validateMinStabilitySpread(i interface{}) error {
	v, ok := i.(osmomath.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min spread must be positive or zero: %s", v)
	}

	if v.GT(osmomath.OneDec()) {
		return fmt.Errorf("min spread is too large: %s", v)
	}

	return nil
}
