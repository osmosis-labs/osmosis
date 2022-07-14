package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The first block whose timestamp is after the duration is counted as the end of the epoch
func (k Keeper) AfterEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.AfterEpochEnd(ctx, identifier, epochNumber)
}

// New epoch is next block of epoch end block
func (k Keeper) BeforeEpochStart(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.BeforeEpochStart(ctx, identifier, epochNumber)
}
