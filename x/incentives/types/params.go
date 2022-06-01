package types

import (
	epochtypes "github.com/osmosis-labs/osmosis/v9/x/epochs/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(distrEpochIdentifier string) Params {
	return Params{
		DistrEpochIdentifier: distrEpochIdentifier,
	}
}

// default minting module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier: "week",
	}
}

// validate params.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}
	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
	}
}
