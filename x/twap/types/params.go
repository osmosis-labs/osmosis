package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	epochtypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
)

// Parameter store keys.
var (
	KeyPruneEpochIdentifier = []byte("PruneEpochIdentifier")
)

const defaultPruneEpochIdentifier = "day"

// ParamTable for twap module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(pruneEpochIdentifier string) Params {
	return Params{
		PruneEpochIdentifier: pruneEpochIdentifier,
	}
}

// default twap module parameters.
func DefaultParams() Params {
	return Params{
		PruneEpochIdentifier: defaultPruneEpochIdentifier,
	}
}

// validate params.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierString(p.PruneEpochIdentifier); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPruneEpochIdentifier, &p.PruneEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
	}
}
