package types

func NewGenesisState(farms []Farm, farmers []Farmer, historicalRecords []GenesisHistoricalRecord) *GenesisState {
	return &GenesisState{
		Farms:             farms,
		Farmers:           farmers,
		HistoricalRecords: historicalRecords,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(nil, nil, nil)
}
