package types

import (
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Incentives parameters key store.
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams takes an epoch distribution identifier, then returns an incentives Params struct.
func NewParams(distrEpochIdentifier string) Params {
	return Params{
		DistrEpochIdentifier: distrEpochIdentifier,
	}
}

// DefaultParams returns the default incentives module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier: "week",
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs takes the parameter struct and associates the paramsubspace key and field of the parameters as a KVStore.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
	}
}
