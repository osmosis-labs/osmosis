package types

type TokenPair struct {
	TokenA string
	TokenB string
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}

func init() {
	// no-op
}
