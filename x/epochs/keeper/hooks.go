package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) OnEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.OnEpochEnd(ctx, identifier, epochNumber)
}

func (k Keeper) OnEpochStart(ctx sdk.Context, identifier string, epochNumber int64) {
	k.hooks.OnEpochStart(ctx, identifier, epochNumber)
}
