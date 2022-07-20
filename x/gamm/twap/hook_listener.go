package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

var _ types.GammHooks = &gammhook{}
var _ epochtypes.EpochHooks = &epochhook{}

type epochhook struct {
	k Keeper
}

func (hook *epochhook) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	if epochIdentifier == hook.k.PruneEpochIdentifier(ctx) {
		hook.k.pruneOldTwaps(ctx)
	}
}

func (hook *epochhook) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {}

type gammhook struct {
	k Keeper
}

func (k Keeper) GammHooks() types.GammHooks {
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

func (hook *gammhook) BeforeSwap(ctx sdk.Context, poolId uint64) {
	err := hook.k.updateTwapIfNotRedundant(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

func (hook *gammhook) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
}

func (hook *gammhook) BeforeJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	err := hook.k.updateTwapIfNotRedundant(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

func (hook *gammhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
}

func (hook *gammhook) AfterExitPool(_ sdk.Context, _ sdk.AccAddress, _ uint64, _ sdk.Int, _ sdk.Coins) {
}
