package types

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
func Validate(gs *GenesisState) error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}
