package types

var AtomDenomination string = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
var OsmosisDenomination string = "uosmo"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Routes: []SearcherRoutes{},
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

type TokenPair struct {
	TokenA string
	TokenB string
}

// Routes entered into the genesis state must start and end with the same denomination and
// the denomination must be Osmo or Atom
func (gs GenesisState) CheckRoutes() error {
	seenTokenPairs := make(map[TokenPair]bool)
	for _, searcherRoutes := range gs.Routes {
		// Validate the searcherRoutes
		if err := searcherRoutes.Validate(); err != nil {
			return err
		}

		// sort token pair
		tokenPair := TokenPair{
			TokenA: searcherRoutes.TokenA,
			TokenB: searcherRoutes.TokenB,
		}
		if tokenPair.TokenA > tokenPair.TokenB {
			tokenPair.TokenA, tokenPair.TokenB = tokenPair.TokenB, tokenPair.TokenA
		}

		// Validate that the token pair is unique
		if _, ok := seenTokenPairs[tokenPair]; ok {
			return ErrDuplicateTokenPair
		}

		seenTokenPairs[tokenPair] = true
	}

	return nil
}
