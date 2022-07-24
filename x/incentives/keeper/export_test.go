package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/incentives/types"
)

// AddGaugeRefByKey appends the provided gauge ID into an array associated with the provided key.
func (k Keeper) AddGaugeRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.addGaugeRefByKey(ctx, key, guageID)
}

// DeleteGaugeRefByKey removes the provided gauge ID from an array associated with the provided key.
func (k Keeper) DeleteGaugeRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.deleteGaugeRefByKey(ctx, key, guageID)
}

// GetGaugeRefs returns the gauge IDs specified by the provided key.
func (k Keeper) GetGaugeRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getGaugeRefs(ctx, key)
}

// GetAllGaugeIDsByDenom returns all active gauge-IDs associated with lockups of the provided denom.
func (k Keeper) GetAllGaugeIDsByDenom(ctx sdk.Context, denom string) []uint64 {
	return k.getAllGaugeIDsByDenom(ctx, denom)
}

// MoveUpcomingGaugeToActiveGauge moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) MoveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge types.Gauge) error {
	return k.moveUpcomingGaugeToActiveGauge(ctx, gauge)
}

// MoveActiveGaugeToFinishedGauge moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) MoveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge types.Gauge) error {
	return k.moveActiveGaugeToFinishedGauge(ctx, gauge)
}

// ChargeFeeIfSufficientFeeDenomBalance see chargeFeeIfSufficientFeeDenomBalance spec.
func (k Keeper) ChargeFeeIfSufficientFeeDenomBalance(ctx sdk.Context, address sdk.AccAddress, fee int64, gaugeCoins sdk.Coins) error {
	return k.chargeFeeIfSufficientFeeDenomBalance(ctx, address, fee, gaugeCoins)
}
