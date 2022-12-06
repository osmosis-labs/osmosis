package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

func (g *GenesisState) Validate() error {
	return nil
}
