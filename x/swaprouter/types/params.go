package types

import (
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamTable for swaprouter module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(pruneEpochIdentifier string, recordHistoryKeepPeriod time.Duration) Params {
	return Params{}
}

// DefaultParams returns default swaprouter module parameters.
func DefaultParams() Params {
	return Params{}
}

// Validate validate params.
func (p Params) Validate() error {
	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validate genesis state.
func (g GenesisState) Validate() error {
	return g.Params.Validate()
}
