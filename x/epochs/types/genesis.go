package types

import (
	"errors"
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
	// TODO: Epochs identifiers should be unique
	for _, epoch := range gs.Epochs {
		if epoch.Identifier == "" {
			return errors.New("epoch identifier should NOT be empty")
		}
		if epoch.Duration == 0 {
			return errors.New("epoch duration should NOT be 0")
		}
	}
	return nil
}
