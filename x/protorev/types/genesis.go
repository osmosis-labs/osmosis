package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var AtomDenomination string = "ATOM"
var OsmosisDenomination string = "OSMO"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// Validate entered routes
	if err := gs.CheckRoutes(); err != nil {
		return err
	}
	return gs.Params.Validate()
}

// Routes entered into the genesis state must start and end with the same denomination and
// the denomination must be Osmo or Atom
func (gs GenesisState) CheckRoutes() error {
	for _, route := range gs.Routes {
		// The arb denomination must be tradable
		if route.ArbDenom != AtomDenomination && route.ArbDenom != OsmosisDenomination {
			return sdkerrors.Wrapf(ErrInvalidArbDenom, "entered denomination was %s but only %s and %s are allowed", route.ArbDenom, AtomDenomination, OsmosisDenomination)
		}

		uniquePools := make(map[uint64]bool)
		for _, pool := range route.Pools {
			uniquePools[pool] = true
		}

		// There must be at least three pools hit for it to be a valid route
		if len(uniquePools) < 3 {
			return sdkerrors.Wrapf(ErrInvalidRoute, "the length of the entered cyclic arbitrage route must hit at least three pools: entered number of pools %d", len(uniquePools))
		}
	}

	return nil
}
