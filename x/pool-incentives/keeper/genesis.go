package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

<<<<<<< HEAD:x/pool-incentives/genesis.go
	"github.com/osmosis-labs/osmosis/v10/x/pool-incentives/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/pool-incentives/types"
=======
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
>>>>>>> 61a207f8 (chore: move init export genesis to keepers (#1631)):x/pool-incentives/keeper/genesis.go
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	k.SetLockableDurations(ctx, genState.LockableDurations)
	if genState.DistrInfo == nil {
		k.SetDistrInfo(ctx, types.DistrInfo{
			TotalWeight: sdk.NewInt(0),
			Records:     nil,
		})
	} else {
		k.SetDistrInfo(ctx, *genState.DistrInfo)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	distrInfo := k.GetDistrInfo(ctx)

	return &types.GenesisState{
		Params:            k.GetParams(ctx),
		LockableDurations: k.GetLockableDurations(ctx),
		DistrInfo:         &distrInfo,
	}
}
