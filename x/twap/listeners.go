package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

var (
	_ gammtypes.GammHooks                  = &swaprouterhook{}
	_ epochtypes.EpochHooks                = &epochhook{}
	_ swaproutertypes.PoolCreationListener = &swaprouterhook{}
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

type swaprouterhook struct {
	k Keeper
}

func (k Keeper) GammHooks() gammtypes.GammHooks {
	return &swaprouterhook{k}
}

func (k Keeper) PoolCreationListeners() swaproutertypes.PoolCreationListener {
	return &swaprouterhook{k}
}

// AfterPoolCreated is called after CreatePool
func (hook *swaprouterhook) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := hook.k.afterCreatePool(ctx, poolId)
	// Will halt pool creation
	if err != nil {
		panic(err)
	}
}

func (hook *swaprouterhook) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *swaprouterhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *swaprouterhook) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}
