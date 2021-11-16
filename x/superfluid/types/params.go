package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// Parameter store keys
var (
	KeyRefreshEpochIdentifier = []byte("RefreshEpochIdentifier")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(refreshEpochIdentifier string) Params {
	return Params{
		RefreshEpochIdentifier: refreshEpochIdentifier,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		RefreshEpochIdentifier: "day",
	}
}

// validate params
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.RefreshEpochIdentifier); err != nil {
		return err
	}
	return nil
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyRefreshEpochIdentifier, &p.RefreshEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
	}
}
