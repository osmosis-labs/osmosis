package types

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// TODO: validate each asset's source chain and denom
	err := gs.Params.Validate()
	if err != nil {
		return err
	}
	return nil
}
