package lockup

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/osmosis-labs/osmosis/v30/x/lockup/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called on every block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
}

var (
	numLocksToDelete = 1_000
)

// Called every block to automatically unlock matured locks.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	if ctx.BlockHeight()%120 == 0 {
		// TODO: Change this logic to "know" when the next unbonding time is, and only unlock at that time.
		// At each unbond, do an iterate to find the next unbonding time and wait until then.
		// delete synthetic locks matured before lockup deletion
		k.DeleteAllMaturedSyntheticLocks(ctx)

		// withdraw and delete locks. Requires the corresponding synthetic locks to be deleted.
		// This is guaranteed, as we delete _ALL_ synthetic locks before withdrawing a bounded number of native locks.
		k.WithdrawMaturedLocks(ctx, numLocksToDelete)
	}
	return []abci.ValidatorUpdate{}
}

// TODO: add invariant that no native lockup existent synthetic lockup exists by calling GetAllSyntheticLockups
// TODO: if superfluid does not delete synthetic lockup before native lockup deletion, it won't be able to be deleted
