package tradingtiers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	txfeestypes "github.com/osmosis-labs/osmosis/v25/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	// Determine the current osmo usd value

	// Set this value in the store and

	// Get the current epoch number

	// Iterate over current epoch - 1 for AccountDailyOsmoVolumePrefix
	// For each entry, use the current osmo usd value to determine the usd volume the account made.
	// Add this volume to the value of the respective AccountRollingWindowUSDVolumePrefix entry.
	// If this summation results in a tier increase, change the key accordingly.

	// Iterate over current epoch - 30 for AccountDailyOsmoVolumePrefix
	// For each entry, use the current osmo usd value to determine the usd volume the account made.
	// Subtract this volume from the value of the respective AccountRollingWindowUSDVolumePrefix entry.
	// If this subtraction results in a tier decrease, change the key accordingly.
	// If the subtraction resulkts in a zero value, delete the key.

	// Set the cached value for the epoch number
	return nil
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return txfeestypes.ModuleName
}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
