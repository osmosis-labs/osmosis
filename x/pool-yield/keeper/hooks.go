package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
)

type Hooks struct {
	k Keeper
}

var _ gammtypes.GammHooks = Hooks{}

// Create new pool yield hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// creates a farm for each pool’s lockable duration
func (h Hooks) AfterPoolCreated(ctx sdk.Context, poolId uint64) {
	h.k.CreatePoolFarms(ctx, poolId)
}
