package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

var (
	_ gammtypes.GammHooks   = &gammhook{}
	_ epochtypes.EpochHooks = &epochhook{}
)

type epochhook struct {
	k Keeper
}

func (k Keeper) EpochHooks() epochtypes.EpochHooks {
	return &epochhook{k}
}

// GetModuleName implements types.EpochHooks.
func (*epochhook) GetModuleName() string {
	return types.ModuleName
}

func (hook *epochhook) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier == hook.k.PruneEpochIdentifier(ctx) {
		lastKeptTime := ctx.BlockTime().Add(-hook.k.RecordHistoryKeepPeriod(ctx))
		poolIdToStartFrom := hook.k.poolmanagerKeeper.GetNextPoolId(ctx) - 1
		if poolIdToStartFrom > 0 {
			hook.k.SetPruningState(ctx, types.PruningState{
				IsPruning:      true,
				LastKeptTime:   lastKeptTime,
				LastSeenPoolId: poolIdToStartFrom,
			})
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

// AfterCFMMPoolCreated is called after CreatePool run on a CFMM pool from x/gamm.
func (hook *gammhook) AfterCFMMPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	hook.k.mustTrackCreatedPool(ctx, poolId)
}

// AfterCFMMSwap is called after SwapExactAmountIn and SwapExactAmountOut in x/gamm.
func (hook *gammhook) AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount osmomath.Int) {
	hook.k.trackChangedPool(ctx, poolId)
}

func (hook *gammhook) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount osmomath.Int, exitCoins sdk.Coins) {
	hook.k.trackChangedPool(ctx, poolId)
}

type concentratedLiquidityListener struct {
	k Keeper
}

func (k Keeper) ConcentratedLiquidityListener() concentratedliquiditytypes.ConcentratedLiquidityListener {
	return &concentratedLiquidityListener{k}
}

func (l *concentratedLiquidityListener) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	l.k.mustTrackCreatedPool(ctx, poolId)
}

func (l *concentratedLiquidityListener) AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	l.k.trackChangedPool(ctx, poolId)
}

func (l *concentratedLiquidityListener) AfterLastPoolPositionRemoved(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	l.k.trackChangedPool(ctx, poolId)
}

func (l *concentratedLiquidityListener) AfterConcentratedPoolSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	l.k.trackChangedPool(ctx, poolId)
}
