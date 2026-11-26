package superfluid

import (
	"github.com/osmosis-labs/osmosis/v32/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v32/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper) {
	epochIdentifier := k.GetEpochIdentifier(ctx)
	numBlocksSinceEpochStart, err := ek.NumBlocksSinceEpochStart(ctx, epochIdentifier)
	if err != nil {
		// The epoch identifier is an x/incentives param, and nothing ties it to an epoch that x/epochs knows.
		// Not finding the epoch must not halt the chain: the epoch logic is skipped until the param is corrected.
		ctx.Logger().Error("superfluid: skipping epoch logic, could not get the number of blocks since epoch start",
			"epoch_identifier", epochIdentifier, "error", err)
		return
	}
	if numBlocksSinceEpochStart == 0 {
		k.AfterEpochStartBeginBlock(ctx)
	}
}
