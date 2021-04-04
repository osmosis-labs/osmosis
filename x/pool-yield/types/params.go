package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMintDenom         = []byte("PoolYieldMintDenom")
	KeyLockableDurations = []byte("PoolYieldLockableDurations")
)

// ParamKeyTable for pool-yield module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the pool-yield module
func NewParams(mintDenom string, lockableDurations []time.Duration) Params {
	return Params{
		MintDenom:         mintDenom,
		LockableDurations: lockableDurations,
	}
}

// DefaultParams is the default parameter configuration for the pool-yield module
func DefaultParams() Params {
	return Params{
		MintDenom:         "osmo",
		LockableDurations: []time.Duration{},
	}
}

// Validate all pool-yield module parameters
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	return validateLockableDurations(p.LockableDurations)
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateLockableDurations(i interface{}) error {
	_, ok := i.([]time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyLockableDurations, &p.LockableDurations, validateLockableDurations),
	}
}
