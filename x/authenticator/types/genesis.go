package types

// DefaultIndex is the default global index
const DefaultIndex uint64 = 0

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		NextAuthenticatorId: DefaultIndex,
		AuthenticatorData:   []AuthenticatorData{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
