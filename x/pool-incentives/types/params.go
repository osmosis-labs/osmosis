package types

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var KeyMintedDenom = []byte("MintedDenom")

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(mintedDenom string) Params {
	return Params{
		MintedDenom: mintedDenom,
	}
}

// DefaultParams is the default parameter configuration for the pool-incentives module.
func DefaultParams() Params {
	return NewParams(sdk.DefaultBondDenom)
}

func (p Params) Validate() error {
	if err := validateMintedDenom(p.MintedDenom); err != nil {
		return err
	}
	return nil
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

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintedDenom, &p.MintedDenom, validateMintedDenom),
	}
}
