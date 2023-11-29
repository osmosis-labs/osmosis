package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// At the end of each epoch, set the denom pair routes.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	// _ = osmoutils.ApplyFuncIfNoError(ctx, func(cacheCtx sdk.Context) error {
	// 	_, err := k.SetDenomPairRoutes(cacheCtx)
	// 	return err
	// })

	return nil
}

// Hooks wrapper struct for poolmanager keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

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
