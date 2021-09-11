package lockup

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetLastLockID(ctx, genState.LastLockId)
	for i, lock := range genState.Locks {
		if i%10000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("lock number %d, entry %d\n", lock.ID, i))
		}
		// reset lock's main operation is to store reference queues for iteration
		if err := k.ResetLock(ctx, lock); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	locks, err := k.GetPeriodLocks(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		LastLockId: k.GetLastLockID(ctx),
		Locks:      locks,
	}
}
