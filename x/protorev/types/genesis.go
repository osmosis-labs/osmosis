package types

var AtomDenomination string = "ATOM"
var OsmosisDenomination string = "OSMO"

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

// Routes entered into the genesis state must start and end with the same denomination and
// the denomination must be Osmo or Atom
func (gs GenesisState) CheckRoutes() error {
	for _, searcherRoutes := range gs.Routes {
		// Validate the searcherRoutes
		if err := searcherRoutes.Validate(); err != nil {
			return err
		}
	}

	return nil
}
