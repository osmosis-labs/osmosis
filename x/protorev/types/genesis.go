package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InputAmountList contains a list of input amounts to test when creating the optimal amount to swap in the
// binary search method
var InputAmountList []sdk.Int

// Denoms stored in order to build the routes
var AtomDenomination string = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
var OsmosisDenomination string = "uosmo"

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
	// Init all of the input amounts
	InputAmountList = make([]sdk.Int, 0)
	for i := 0; i < 1000000000; i += 100000 {
		InputAmountList = append(InputAmountList, sdk.NewInt(int64(i)))
	}
}

// Routes entered into the genesis state must start and end with the same denomination and
// the denomination must be Osmo or Atom
func (gs GenesisState) CheckRoutes() error {
	seenTokenPairs := make(map[TokenPair]bool)
	for _, tokenPairArbRoutes := range gs.TokenPairs {
		// Validate the arb routes
		if err := tokenPairArbRoutes.Validate(); err != nil {
			return err
		}

		tokenPair := TokenPair{
			TokenA: tokenPairArbRoutes.TokenA,
			TokenB: tokenPairArbRoutes.TokenB,
		}
		// Validate that the token pair is unique
		if _, ok := seenTokenPairs[tokenPair]; ok {
			return ErrDuplicateTokenPair
		}

		seenTokenPairs[tokenPair] = true
	}

	return nil
}
