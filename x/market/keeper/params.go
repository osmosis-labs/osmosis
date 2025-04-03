package keeper

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/market/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MinStabilitySpread is the minimum spread applied to swaps to / from Note.
// Intended to prevent swing trades exploiting oracle period delays
func (k Keeper) MinStabilitySpread(ctx sdk.Context) (res osmomath.Dec) {
	k.paramSpace.Get(ctx, types.KeyMinStabilitySpread, &res)
	return
}

// PoolRecoveryPeriod is the period required to recover Symphony&Note Pools to the MintBasePool & BurnBasePool
func (k Keeper) PoolRecoveryPeriod(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyPoolRecoveryPeriod, &res)
	return
}

// GetParams returns the total set of market parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSetIfExists(ctx, &params)
	params.ExchangePool = k.GetExchangePoolBalance(ctx).Amount.ToLegacyDec()
	return params
}

// SetParams sets the total set of market parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
