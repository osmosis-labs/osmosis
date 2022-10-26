package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

var (
	_ gammtypes.GammHooks                  = &gammhook{}
	_ epochtypes.EpochHooks                = &epochhook{}
	_ swaproutertypes.PoolCreationListener = &gammhook{}
)

type epochhook struct {
	k Keeper
}

func (k Keeper) EpochHooks() epochtypes.EpochHooks {
	return &epochhook{k}
}

func (hook *epochhook) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier == hook.k.PruneEpochIdentifier(ctx) {
		if err := hook.k.pruneRecords(ctx); err != nil {
			ctx.Logger().Error("Error pruning old twaps at the epoch end", err)
		}
	}
	return nil
}

func (hook *epochhook) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

type gammhook struct {
	k Keeper
}

func (k Keeper) GammHooks() gammtypes.GammHooks {
	return &gammhook{k}
}

func (k Keeper) PoolCreationListeners() swaproutertypes.PoolCreationListener {
	return &gammhook{k}
}

// AfterPoolCreated is called after CreatePool
func (hook *gammhook) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := hook.k.afterCreatePool(ctx, poolId)
	// Will halt pool creation
	if err != nil {
		panic(err)
	}
}

func (hook *gammhook) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}
