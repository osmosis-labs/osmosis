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
		ModuleAccountBalance: sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt()),
		StartTime:            time.Now(),
		DurationUntilDecay:   DefaultDurationUntilDecay, // 1 month
		DurationOfDecay:      DefaultDurationOfDecay,    // 5 months
		InitialClaimables:    []banktypes.Balance{},
		Activities:           []UserActions{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
