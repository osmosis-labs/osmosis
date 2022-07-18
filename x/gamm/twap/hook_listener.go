package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/v10/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/types"
)

var _ types.GammHooks = &gammhook{}
var _ epochtypes.EpochHooks = &epochhook{}

type epochhook struct {
	k twapkeeper
}

func (hook *epochhook) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// TODO:
	// if epochIdentifier == hook.k.PruneIdentifier() {
	//	 hook.k.pruneOldTwaps(ctx)
	// }
}

func (hook *epochhook) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {}

type gammhook struct {
	k twapkeeper
}

// AfterPoolCreated is called after CreatePool
func (hook *gammhook) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	// TODO: Log pool creation to begin creating TWAPs for it
}

func (hook *gammhook) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	// Log that this pool had a potential spot price change
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	// Log that this pool had a potential spot price change
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterExitPool(_ sdk.Context, _ sdk.AccAddress, _ uint64, _ sdk.Int, _ sdk.Coins) {
}
