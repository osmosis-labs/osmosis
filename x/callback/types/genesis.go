package types

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(
	params Params,
	callbacks []Callback,
) *GenesisState {
	var callbacksCopy []*Callback
	for _, c := range callbacks {
		callbacksCopy = append(callbacksCopy, &c)
	}
	return &GenesisState{
		Params:    params,
		Callbacks: callbacksCopy,
	}
}

// DefaultGenesis returns a default genesis state.
func DefaultGenesis() *GenesisState {
	defaultParams := DefaultParams()
	return NewGenesisState(defaultParams, nil)
}

// Validate perform object fields validation.
func (g GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}
	for _, callback := range g.GetCallbacks() {
		if err := callback.Validate(); err != nil {
			return err
		}
	}
	return nil
}
