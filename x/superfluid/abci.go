package superfluid

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper) {
	numBlocksSinceEpochStart, err := ek.NumBlocksSinceEpochStart(ctx, k.GetEpochIdentifier(ctx))
	if err != nil {
		panic(err)
	}
	if numBlocksSinceEpochStart == 0 {
		k.AfterEpochStartBeginBlock(ctx)
	}
}

// Called every block to automatically unlock matured locks.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
