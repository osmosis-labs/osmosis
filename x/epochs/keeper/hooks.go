package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AfterEpochEnd gets called at the end of the epoch, end of epoch is the timestamp of first block produced after epoch duration.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) error {
	return k.hooks.AfterEpochEnd(ctx, identifier, epochNumber)
}

// BeforeEpochStart new epoch is next block of epoch end block
func (k Keeper) BeforeEpochStart(ctx sdk.Context, identifier string, epochNumber int64) error {
	return k.hooks.BeforeEpochStart(ctx, identifier, epochNumber)
}
