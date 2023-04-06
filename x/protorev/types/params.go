package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultEnableModule = true
	// Currently configured to be the Skip dev team's address
	// See https://github.com/osmosis-labs/osmosis/issues/4349 for more details
	// Note that governance has full ability to change this live on-chain, and this admin can at most prevent protorev from working.
	// All the settings manager's controls have limits, so it can't lead to a chain halt, excess processing time or prevention of swaps.
	DefaultAdminAccount = "osmo17nv67dvc7f8yr00rhgxd688gcn9t9wvhn783z4"

	ParamStoreKeyEnableModule = []byte("EnableProtoRevModule")
	ParamStoreKeyAdminAccount = []byte("AdminAccount")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(enable bool, admin string) Params {
	return Params{
		Enabled: enable,
		Admin:   admin,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultEnableModule, DefaultAdminAccount)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableModule, &p.Enabled, ValidateBoolean),
		paramtypes.NewParamSetPair(ParamStoreKeyAdminAccount, &p.Admin, ValidateAccount),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if _, err := sdk.AccAddressFromBech32(p.Admin); err != nil {
		return fmt.Errorf("invalid admin account address: %s", p.Admin)
	}

	return nil
}

func ValidateAccount(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if _, err := sdk.AccAddressFromBech32(v); err != nil {
		return fmt.Errorf("invalid account address: %s", v)
	}

	return nil
}

func ValidateBoolean(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
