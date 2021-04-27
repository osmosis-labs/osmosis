package types

import (
	"time"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Epochs: []EpochInfo{
			{
				Identifier:            "weekly",
				StartTime:             time.Time{},
				Duration:              time.Hour * 24 * 7,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
				CurrentEpochEnded:     true,
			},
			{
				Identifier:            "daily",
				StartTime:             time.Time{},
				Duration:              time.Hour * 24,
				CurrentEpoch:          0,
				CurrentEpochStartTime: time.Time{},
				EpochCountingStarted:  false,
				CurrentEpochEnded:     true,
			},
		},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
