package superfluid

import (
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v17/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v17/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, ek types.EpochKeeper) {
	numBlocksSinceEpochStart, err := ek.NumBlocksSinceEpochStart(ctx, k.GetEpochIdentifier(ctx))
	if err != nil {
		panic(err)
	}
	if numBlocksSinceEpochStart == 0 {
		// catch any panics/errors in superfluid, and revert begin block logic if it occurs.
		//nolint:errcheck
		osmoutils.ApplyFuncIfNoError(ctx, func(ctx2 sdk.Context) error {
			k.AfterEpochStartBeginBlock(ctx2)
			return nil
		})
	}
}
