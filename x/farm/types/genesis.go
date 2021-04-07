package types

func NewGenesisState(farms []Farm, farmers []Farmer, historicalEntries []GenesisHistoricalEntry) *GenesisState {
	return &GenesisState{
		Farms:             farms,
		Farmers:           farmers,
		HistoricalEntries: historicalEntries,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(nil, nil, nil)
}
