package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(minter Minter, params Params, reductionStartedEpoch int64, inflationDelta sdk.Dec, developerVestingDelta sdk.Dec) *GenesisState {
	return &GenesisState{
		Minter:                          minter,
		Params:                          params,
		ReductionStartedEpoch:           reductionStartedEpoch,
		InflationTruncationDelta:        inflationDelta,
		DeveloperVestingTruncationDelta: developerVestingDelta,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Minter:                          DefaultInitialMinter(),
		Params:                          DefaultParams(),
		ReductionStartedEpoch:           0,
		InflationTruncationDelta:        sdk.ZeroDec(),
		DeveloperVestingTruncationDelta: sdk.ZeroDec(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return data.Minter.Validate()
}
