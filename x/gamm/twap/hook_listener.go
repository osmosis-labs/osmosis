package twap

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

var (
	_ types.GammHooks       = &gammhook{}
	_ epochtypes.EpochHooks = &epochhook{}
)

type epochhook struct {
	k Keeper
}

func (hook *epochhook) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	if epochIdentifier == hook.k.PruneEpochIdentifier(ctx) {
		fmt.Println("restore logic in subsequent PR")
		// hook.k.pruneOldTwaps(ctx)
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
	// err := hook.k.afterCreatePool(ctx, poolId)
	// // Will halt pool creation
	// if err != nil {
	// 	panic(err)
	// }
}

func (hook *gammhook) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterExitPool(_ sdk.Context, _ sdk.AccAddress, _ uint64, _ sdk.Int, _ sdk.Coins) {
}
