package types

import (
	"fmt"
)

type TokenPair struct {
	TokenA string
	TokenB string
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		TokenPairs: []TokenPairArbRoutes{},
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

func init() {
	// no-op
}

// Routes entered into the genesis state must start and end with the same denomination and
// the denomination must be Osmo or Atom. Additionally, there cannot be duplicate routes (same
// token pairs).
func (gs GenesisState) CheckRoutes() error {
	seenTokenPairs := make(map[TokenPair]bool)
	for _, tokenPairArbRoutes := range gs.TokenPairs {
		// Validate the arb routes
		if err := tokenPairArbRoutes.Validate(); err != nil {
			return err
		}

		tokenPair := TokenPair{
			TokenA: tokenPairArbRoutes.TokenIn,
			TokenB: tokenPairArbRoutes.TokenOut,
		}
		// Validate that the token pair is unique
		if _, ok := seenTokenPairs[tokenPair]; ok {
			return fmt.Errorf("duplicate token pair: %s", tokenPair)
		}

		seenTokenPairs[tokenPair] = true
	}

	return nil
}
