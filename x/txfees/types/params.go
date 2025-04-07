package types

import (
	"fmt"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
)

const DefaultSwapFeesEpochIdentifier = "day"

var KeySwapFeesEpochIdentifier = []byte("SwapFeesEpochIdentifier")

// ParamTable for txfees module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(whitelistedFeeTokenSetters []string) Params {
	return Params{}
}

// DefaultParams are the default txfees module parameters.
func DefaultParams() Params {
	return Params{
		SwapFeesEpochIdentifier: DefaultSwapFeesEpochIdentifier,
	}
}

// validate params.
func (p Params) Validate() error {
	if epochtypes.ValidateEpochIdentifierString(p.SwapFeesEpochIdentifier) != nil {
		return fmt.Errorf("treasury parameter SwapFeesEpochIdentifier must be a valid epoch identifier: %s", p.SwapFeesEpochIdentifier)
	}
	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeySwapFeesEpochIdentifier, &p.SwapFeesEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
	}
}
