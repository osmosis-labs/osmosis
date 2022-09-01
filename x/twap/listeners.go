package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v11/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"
)

var (
	_ types.GammHooks       = &gammhook{}
	_ epochtypes.EpochHooks = &epochhook{}
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

func (k Keeper) GammHooks() types.GammHooks {
	return &gammhook{k}
}

// AfterPoolCreated is called after CreatePool
func (hook *gammhook) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) error {
	return hook.k.afterCreatePool(ctx, poolId)
}

func (hook *gammhook) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) error {
	hook.k.trackChangedPool(ctx, poolId)
	return nil
}

func (hook *gammhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) error {
	hook.k.trackChangedPool(ctx, poolId)
	return nil
}

func (hook *gammhook) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) error {
	hook.k.trackChangedPool(ctx, poolId)
	return nil
}
