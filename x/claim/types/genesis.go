package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Actions []Action

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		AirdropAmount:      sdk.NewInt(0),
		AirdropStart:       time.Now(),
		DurationUntilDecay: time.Hour * 24 * 30,     // 1 month
		DurationOfDecay:    time.Hour * 24 * 30 * 5, // 5 months
		Claimables:         []banktypes.Balance{},
		Activities:         []UserActivities{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
