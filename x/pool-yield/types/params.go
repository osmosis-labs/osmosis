package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyMintedDenom     = []byte("MintedDenom")
	KeyAllocationRatio = []byte("AllocationRatio")
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(mintedDenom string, allocationRatio sdk.Dec) Params {
	return Params{
		MintedDenom:     mintedDenom,
		AllocationRatio: allocationRatio,
	}
}

// DefaultParams is the default parameter configuration for the pool-yield module
func DefaultParams() Params {
	return NewParams("stake", sdk.NewDecWithPrec(2, 1))
}

func (p Params) Validate() error {
	if err := validateMintedDenom(p.MintedDenom); err != nil {
		return err
	}
	return validateAllocationRatio(p.AllocationRatio)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMintedDenom(i interface{}) error {
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

func validateAllocationRatio(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return errors.New("allocation ratio should not be negative")
	}

	if v.GT(sdk.NewDec(1)) {
		return errors.New("allocation ratio should be lesser than 1")
	}

	return nil
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintedDenom, &p.MintedDenom, validateMintedDenom),
		paramtypes.NewParamSetPair(KeyAllocationRatio, &p.AllocationRatio, validateAllocationRatio),
	}
}
