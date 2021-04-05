package pool_yield

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/c-osmosis/osmosis/x/pool-yield/keeper"
)

func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	if ctx.BlockHeight() > 1 {
		asset := k.GetAllocatableAsset(ctx)
		if asset.IsValid() && asset.IsPositive() {
			err := k.AllocateAsset(ctx, asset)
			if err != nil {
				panic(err)
			}
		}
	}
}
