package api

// DefaultGenesis creates a default GenesisState object
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Lbps:          []*LBP{},
		UserPositions: []*UserPosition{},
		NextLbpNumber: 1,
		Params:         DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
