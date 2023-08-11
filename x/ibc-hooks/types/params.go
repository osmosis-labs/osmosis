package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyAsyncAckAllowList = []byte("AsyncAckAllowList")

	_ paramtypes.ParamSet = &Params{}
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(allowedAsyncAckContracts []string) Params {
	return Params{
		AllowedAsyncAckContracts: allowedAsyncAckContracts,
	}
}

// DefaultParams returns default concentrated-liquidity module parameters.
func DefaultParams() Params {
	return Params{
		AllowedAsyncAckContracts: []string{},
	}
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAsyncAckAllowList, &p.AllowedAsyncAckContracts, validateAsyncAckAllowList),
	}
}

// Validate params.
func (p Params) Validate() error {
	if err := validateAsyncAckAllowList(p.AllowedAsyncAckContracts); err != nil {
		return err
	}
	return nil
}

func validateAsyncAckAllowList(i interface{}) error {
	allowedContracts, ok := i.([]string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, contract := range allowedContracts {
		if _, err := sdk.AccAddressFromBech32(contract); err != nil {
			return err
		}
	}

	return nil
}
